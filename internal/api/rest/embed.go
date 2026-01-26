package rest

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFiles embed.FS

// StaticFS returns the embedded static files
func StaticFS() http.FileSystem {
	sub, _ := fs.Sub(staticFiles, "static")
	return http.FS(sub)
}

// ServeStatic serves the embedded dashboard
func ServeStatic(router *gin.Engine) {
	// Serve static files
	router.StaticFS("/app", StaticFS())

	// Redirect root to /app for dashboard
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/app")
	})
}
