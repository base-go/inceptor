package rest

import (
	"github.com/flakerimi/inceptor/internal/auth"
	"github.com/flakerimi/inceptor/internal/core"
	"github.com/flakerimi/inceptor/internal/storage"
	"github.com/gin-gonic/gin"
)

// Server holds the REST API server
type Server struct {
	router      *gin.Engine
	handler     *Handler
	authHandler *AuthHandler
	authManager *auth.Manager
}

// NewServer creates a new REST API server
func NewServer(repo storage.Repository, fileStore storage.FileStore, alerter *core.AlertManager, authManager *auth.Manager, adminKey string) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	handler := NewHandler(repo, fileStore, alerter)
	authHandler := NewAuthHandler(authManager)

	s := &Server{
		router:      router,
		handler:     handler,
		authHandler: authHandler,
		authManager: authManager,
	}

	s.setupRoutes(repo, adminKey)

	return s
}

// setupRoutes configures all routes
func (s *Server) setupRoutes(repo storage.Repository, adminKey string) {
	// Middleware
	s.router.Use(Recovery())
	s.router.Use(CORS())

	// Serve embedded dashboard
	ServeStatic(s.router)

	// Health check (no auth)
	s.router.GET("/health", s.handler.Health)
	s.router.GET("/ready", s.handler.Health)

	// API v1
	v1 := s.router.Group("/api/v1")

	// Auth routes (no auth required)
	v1.GET("/auth/status", s.authHandler.Status)
	v1.POST("/auth/login", s.authHandler.Login)
	v1.POST("/auth/logout", s.authHandler.Logout)

	// Change password (requires valid session)
	v1.POST("/auth/change-password", SessionAuth(s.authManager), s.authHandler.ChangePassword)

	// Public crash submission endpoint (requires app API key)
	v1.POST("/crashes", APIKeyAuth(repo, adminKey), s.handler.SubmitCrash)

	// Authenticated routes
	authenticated := v1.Group("")
	authenticated.Use(APIKeyAuth(repo, adminKey))
	{
		// Crashes
		authenticated.GET("/crashes", s.handler.ListCrashes)
		authenticated.GET("/crashes/:id", s.handler.GetCrash)
		authenticated.DELETE("/crashes/:id", s.handler.DeleteCrash)

		// Groups
		authenticated.GET("/groups", s.handler.ListGroups)
		authenticated.GET("/groups/:id", s.handler.GetGroup)
		authenticated.PATCH("/groups/:id", s.handler.UpdateGroup)

		// App stats (app can access their own stats)
		authenticated.GET("/apps/:id/stats", s.handler.GetAppStats)

		// Alerts
		authenticated.GET("/alerts", s.handler.ListAlerts)
	}

	// Admin-only routes
	admin := v1.Group("")
	admin.Use(APIKeyAuth(repo, adminKey), AdminOnly())
	{
		// App management
		admin.POST("/apps", s.handler.CreateApp)
		admin.GET("/apps", s.handler.ListApps)
		admin.GET("/apps/:id", s.handler.GetApp)

		// Alert management
		admin.POST("/alerts", s.handler.CreateAlert)
		admin.DELETE("/alerts/:id", s.handler.DeleteAlert)
	}
}

// Router returns the Gin router
func (s *Server) Router() *gin.Engine {
	return s.router
}

// Run starts the server
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
