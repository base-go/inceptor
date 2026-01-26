package rest

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed static/*
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
	router.GET("/apps/_payload.json", func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
	router.GET("/crashes/_payload.json", func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
	router.GET("/groups/_payload.json", func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
	router.GET("/settings/_payload.json", func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
