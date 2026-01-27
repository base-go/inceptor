package rest

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:static
var staticFiles embed.FS

// ServeStatic serves the embedded dashboard at root
func ServeStatic(router *gin.Engine) {
	staticFS, _ := fs.Sub(staticFiles, "static")
	fileServer := http.FileServer(http.FS(staticFS))

	// Serve static assets directly
	router.GET("/_nuxt/*filepath", func(c *gin.Context) {
		c.Request.URL.Path = "/_nuxt" + c.Param("filepath")
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	router.GET("/_payload.json", func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// SPA routes - serve index.html for these paths
	spaRoutes := []string{"/", "/apps", "/crashes", "/crashes/:id", "/groups", "/settings"}
	for _, route := range spaRoutes {
		route := route // capture
		router.GET(route, func(c *gin.Context) {
			// Check if there's a specific HTML file for this route
			path := strings.TrimPrefix(c.Request.URL.Path, "/")
			if path == "" {
				path = "index.html"
			} else {
				path = path + "/index.html"
			}

			// Try to serve the specific page, fallback to index.html
			if data, err := staticFiles.ReadFile("static/" + path); err == nil {
				c.Data(http.StatusOK, "text/html; charset=utf-8", data)
			} else if data, err := staticFiles.ReadFile("static/index.html"); err == nil {
				c.Data(http.StatusOK, "text/html; charset=utf-8", data)
			} else {
				c.String(http.StatusNotFound, "Not found")
			}
		})
	}

	// Serve payload.json for each route
	payloadRoutes := map[string]string{
		"/apps/_payload.json":     "static/apps/_payload.json",
		"/crashes/_payload.json":  "static/crashes/_payload.json",
		"/groups/_payload.json":   "static/groups/_payload.json",
		"/settings/_payload.json": "static/settings/_payload.json",
	}
	for route, filePath := range payloadRoutes {
		filePath := filePath // capture
		router.GET(route, func(c *gin.Context) {
			if data, err := staticFiles.ReadFile(filePath); err == nil {
				c.Data(http.StatusOK, "application/json", data)
			} else {
				c.String(http.StatusNotFound, "Not found")
			}
		})
	}
}
