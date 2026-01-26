package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/flakerimi/inceptor/internal/core"
	"github.com/flakerimi/inceptor/internal/storage"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Note: This file contains the gRPC server implementation.
// The proto file needs to be compiled with protoc to generate the Go code.
// For now, we'll define the interfaces and implementation manually.

// CrashServiceServer is the gRPC server interface
type CrashServiceServer interface {
	SubmitCrash(context.Context, *CrashReport) (*CrashResponse, error)
	SubmitCrashBatch(context.Context, *CrashBatchRequest) (*CrashBatchResponse, error)
	SubmitCrashStream(CrashService_SubmitCrashStreamServer) error
	GetCrash(context.Context, *GetCrashRequest) (*CrashReport, error)
	ListCrashes(context.Context, *ListCrashesRequest) (*ListCrashesResponse, error)
	ListCrashesStream(*ListCrashesRequest, CrashService_ListCrashesStreamServer) error
}

// Server implements the gRPC crash service
type Server struct {
	repo      storage.Repository
	fileStore storage.FileStore
	grouper   *core.Grouper
	alerter   *core.AlertManager
	adminKey  string
}

// NewServer creates a new gRPC server
func NewServer(repo storage.Repository, fileStore storage.FileStore, alerter *core.AlertManager, adminKey string) *Server {
	return &Server{
		repo:      repo,
		fileStore: fileStore,
		grouper:   core.NewGrouper(),
		alerter:   alerter,
		adminKey:  adminKey,
	}
}

// Run starts the gRPC server
func (s *Server) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(s.authInterceptor),
		grpc.StreamInterceptor(s.streamAuthInterceptor),
	)

	// Register service (would use generated code in production)
	// pb.RegisterCrashServiceServer(grpcServer, s)

	log.Info().Str("addr", addr).Msg("Starting gRPC server")
	return grpcServer.Serve(lis)
}

// authInterceptor handles authentication for unary calls
func (s *Server) authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Extract API key from metadata
	app, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	// Add app to context
	ctx = context.WithValue(ctx, "app", app)
	return handler(ctx, req)
}

// streamAuthInterceptor handles authentication for streaming calls
func (s *Server) streamAuthInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Extract API key from metadata
	_, err := s.authenticate(ss.Context())
	if err != nil {
		return err
	}

	return handler(srv, ss)
}

// authenticate validates the API key and returns the app
func (s *Server) authenticate(ctx context.Context) (*core.App, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	apiKeys := md.Get("x-api-key")
	if len(apiKeys) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing API key")
	}

	apiKey := apiKeys[0]

	// Check admin key
	if s.adminKey != "" && apiKey == s.adminKey {
		return &core.App{ID: "admin", Name: "Admin"}, nil
	}

	// Hash and lookup
	keyHash := hashAPIKey(apiKey)
	app, err := s.repo.GetAppByAPIKey(ctx, keyHash)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to validate API key")
	}

	if app == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid API key")
	}

	return app, nil
}

// SubmitCrash handles a single crash submission
func (s *Server) SubmitCrash(ctx context.Context, req *CrashReport) (*CrashResponse, error) {
	app := ctx.Value("app").(*core.App)

	crash := protoToCrash(req)
	crash.ID = uuid.New().String()
	crash.AppID = app.ID
	crash.CreatedAt = time.Now().UTC()

	if crash.Environment == "" {
		crash.Environment = core.EnvironmentProduction
	}

	// Generate fingerprint
	crash.Fingerprint = s.grouper.GenerateFingerprint(crash)
	crash.GroupID = uuid.New().String()

	// Get or create group
	group, isNewGroup, err := s.repo.GetOrCreateGroup(ctx, crash)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to process crash group")
	}
	crash.GroupID = group.ID

	// Save to file store
	if logPath, err := s.fileStore.SaveCrashLog(ctx, crash); err == nil {
		crash.LogFilePath = logPath
	}

	// Save to database
	if err := s.repo.CreateCrash(ctx, crash); err != nil {
		return nil, status.Error(codes.Internal, "failed to save crash")
	}

	// Send alert
	if s.alerter != nil {
		eventType := core.AlertEventNewCrash
		if isNewGroup {
			eventType = core.AlertEventNewGroup
		}
		s.alerter.Notify(core.AlertEvent{
			Type:       eventType,
			AppID:      app.ID,
			Crash:      crash,
			Group:      group,
			IsNewGroup: isNewGroup,
		})
	}

	return &CrashResponse{
		Id:         crash.ID,
		GroupId:    crash.GroupID,
		Fingerprint: crash.Fingerprint,
		IsNewGroup: isNewGroup,
	}, nil
}

// SubmitCrashBatch handles batch crash submission
func (s *Server) SubmitCrashBatch(ctx context.Context, req *CrashBatchRequest) (*CrashBatchResponse, error) {
	var results []*CrashResponse
	accepted := 0
	rejected := 0

	for _, crashReport := range req.Crashes {
		resp, err := s.SubmitCrash(ctx, crashReport)
		if err != nil {
			rejected++
			continue
		}
		accepted++
		results = append(results, resp)
	}

	return &CrashBatchResponse{
		Accepted: int32(accepted),
		Rejected: int32(rejected),
		Results:  results,
	}, nil
}

// SubmitCrashStream handles streaming crash submission
func (s *Server) SubmitCrashStream(stream CrashService_SubmitCrashStreamServer) error {
	accepted := 0
	rejected := 0
	var results []*CrashResponse

	for {
		crashReport, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&CrashBatchResponse{
				Accepted: int32(accepted),
				Rejected: int32(rejected),
				Results:  results,
			})
		}
		if err != nil {
			return err
		}

		resp, err := s.SubmitCrash(stream.Context(), crashReport)
		if err != nil {
			rejected++
			continue
		}
		accepted++
		results = append(results, resp)
	}
}

// GetCrash retrieves a single crash
func (s *Server) GetCrash(ctx context.Context, req *GetCrashRequest) (*CrashReport, error) {
	crash, err := s.repo.GetCrash(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve crash")
	}
	if crash == nil {
		return nil, status.Error(codes.NotFound, "crash not found")
	}

	// Load full data from file
	if crash.LogFilePath != "" {
		if fullCrash, err := s.fileStore.GetCrashLog(ctx, crash.LogFilePath); err == nil && fullCrash != nil {
			crash = fullCrash
		}
	}

	return crashToProto(crash), nil
}

// ListCrashes lists crashes
func (s *Server) ListCrashes(ctx context.Context, req *ListCrashesRequest) (*ListCrashesResponse, error) {
	app := ctx.Value("app").(*core.App)

	filter := storage.CrashFilter{
		AppID:       app.ID,
		GroupID:     req.GroupId,
		Platform:    req.Platform,
		Environment: req.Environment,
		ErrorType:   req.ErrorType,
		UserID:      req.UserId,
		Search:      req.Search,
		Limit:       int(req.Limit),
		Offset:      int(req.Offset),
	}

	if req.FromDate != nil {
		t := req.FromDate.AsTime()
		filter.FromDate = &t
	}
	if req.ToDate != nil {
		t := req.ToDate.AsTime()
		filter.ToDate = &t
	}

	crashes, total, err := s.repo.ListCrashes(ctx, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list crashes")
	}

	protosCrashes := make([]*CrashReport, len(crashes))
	for i, c := range crashes {
		protosCrashes[i] = crashToProto(c)
	}

	return &ListCrashesResponse{
		Crashes: protosCrashes,
		Total:   int32(total),
	}, nil
}

// ListCrashesStream streams crashes
func (s *Server) ListCrashesStream(req *ListCrashesRequest, stream CrashService_ListCrashesStreamServer) error {
	resp, err := s.ListCrashes(stream.Context(), req)
	if err != nil {
		return err
	}

	for _, crash := range resp.Crashes {
		if err := stream.Send(crash); err != nil {
			return err
		}
	}

	return nil
}

// Helper functions for converting between core types and proto types

func protoToCrash(p *CrashReport) *core.Crash {
	crash := &core.Crash{
		ID:           p.Id,
		AppID:        p.AppId,
		AppVersion:   p.AppVersion,
		Platform:     p.Platform,
		OSVersion:    p.OsVersion,
		DeviceModel:  p.DeviceModel,
		ErrorType:    p.ErrorType,
		ErrorMessage: p.ErrorMessage,
		Fingerprint:  p.Fingerprint,
		GroupID:      p.GroupId,
		UserID:       p.UserId,
		Environment:  p.Environment,
	}

	if p.CreatedAt != nil {
		crash.CreatedAt = p.CreatedAt.AsTime()
	}

	for _, f := range p.StackTrace {
		crash.StackTrace = append(crash.StackTrace, core.StackFrame{
			FileName:     f.FileName,
			LineNumber:   int(f.LineNumber),
			ColumnNumber: int(f.ColumnNumber),
			MethodName:   f.MethodName,
			ClassName:    f.ClassName,
			Native:       f.Native,
		})
	}

	if p.Metadata != nil {
		crash.Metadata = make(map[string]interface{})
		for k, v := range p.Metadata {
			crash.Metadata[k] = v
		}
	}

	for _, b := range p.Breadcrumbs {
		bc := core.Breadcrumb{
			Type:     b.Type,
			Category: b.Category,
			Message:  b.Message,
			Level:    b.Level,
		}
		if b.Timestamp != nil {
			bc.Timestamp = b.Timestamp.AsTime()
		}
		if b.Data != nil {
			bc.Data = make(map[string]interface{})
			for k, v := range b.Data {
				bc.Data[k] = v
			}
		}
		crash.Breadcrumbs = append(crash.Breadcrumbs, bc)
	}

	return crash
}

func crashToProto(c *core.Crash) *CrashReport {
	p := &CrashReport{
		Id:           c.ID,
		AppId:        c.AppID,
		AppVersion:   c.AppVersion,
		Platform:     c.Platform,
		OsVersion:    c.OSVersion,
		DeviceModel:  c.DeviceModel,
		ErrorType:    c.ErrorType,
		ErrorMessage: c.ErrorMessage,
		Fingerprint:  c.Fingerprint,
		GroupId:      c.GroupID,
		UserId:       c.UserID,
		Environment:  c.Environment,
		CreatedAt:    timestamppb.New(c.CreatedAt),
	}

	for _, f := range c.StackTrace {
		p.StackTrace = append(p.StackTrace, &StackFrame{
			FileName:     f.FileName,
			LineNumber:   int32(f.LineNumber),
			ColumnNumber: int32(f.ColumnNumber),
			MethodName:   f.MethodName,
			ClassName:    f.ClassName,
			Native:       f.Native,
		})
	}

	if c.Metadata != nil {
		p.Metadata = make(map[string]string)
		for k, v := range c.Metadata {
			if s, ok := v.(string); ok {
				p.Metadata[k] = s
			}
		}
	}

	for _, b := range c.Breadcrumbs {
		pb := &Breadcrumb{
			Timestamp: timestamppb.New(b.Timestamp),
			Type:      b.Type,
			Category:  b.Category,
			Message:   b.Message,
			Level:     b.Level,
		}
		if b.Data != nil {
			pb.Data = make(map[string]string)
			for k, v := range b.Data {
				if s, ok := v.(string); ok {
					pb.Data[k] = s
				}
			}
		}
		p.Breadcrumbs = append(p.Breadcrumbs, pb)
	}

	return p
}

func hashAPIKey(apiKey string) string {
	h := sha256.New()
	h.Write([]byte(apiKey))
	return hex.EncodeToString(h.Sum(nil))
}

// Proto message types (would be generated by protoc in production)

type CrashReport struct {
	Id           string
	AppId        string
	AppVersion   string
	Platform     string
	OsVersion    string
	DeviceModel  string
	ErrorType    string
	ErrorMessage string
	StackTrace   []*StackFrame
	Fingerprint  string
	GroupId      string
	UserId       string
	Environment  string
	CreatedAt    *timestamppb.Timestamp
	Metadata     map[string]string
	Breadcrumbs  []*Breadcrumb
}

type StackFrame struct {
	FileName     string
	LineNumber   int32
	ColumnNumber int32
	MethodName   string
	ClassName    string
	Native       bool
}

type Breadcrumb struct {
	Timestamp *timestamppb.Timestamp
	Type      string
	Category  string
	Message   string
	Data      map[string]string
	Level     string
}

type CrashResponse struct {
	Id          string
	GroupId     string
	Fingerprint string
	IsNewGroup  bool
}

type CrashBatchRequest struct {
	Crashes []*CrashReport
}

type CrashBatchResponse struct {
	Accepted int32
	Rejected int32
	Results  []*CrashResponse
}

type GetCrashRequest struct {
	Id string
}

type ListCrashesRequest struct {
	AppId       string
	GroupId     string
	Platform    string
	Environment string
	ErrorType   string
	UserId      string
	FromDate    *timestamppb.Timestamp
	ToDate      *timestamppb.Timestamp
	Search      string
	Limit       int32
	Offset      int32
}

type ListCrashesResponse struct {
	Crashes []*CrashReport
	Total   int32
}

// Stream interfaces (would be generated by protoc)
type CrashService_SubmitCrashStreamServer interface {
	SendAndClose(*CrashBatchResponse) error
	Recv() (*CrashReport, error)
	grpc.ServerStream
}

type CrashService_ListCrashesStreamServer interface {
	Send(*CrashReport) error
	grpc.ServerStream
}
