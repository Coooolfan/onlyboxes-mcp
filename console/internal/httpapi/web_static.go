package httpapi

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	embeddedWebDistDir  = "web_dist"
	embeddedWebIndex    = "index.html"
	webAssetsPathPrefix = "/assets/"
)

//go:embed web_dist
var embeddedWebDist embed.FS

var embeddedWebFS = mustEmbeddedWebFS()

func mustEmbeddedWebFS() fs.FS {
	webFS, err := fs.Sub(embeddedWebDist, embeddedWebDistDir)
	if err != nil {
		panic("failed to load embedded web dist: " + err.Error())
	}
	return webFS
}

func registerEmbeddedWebRoutes(router *gin.Engine) {
	fileServer := http.FileServer(http.FS(embeddedWebFS))

	router.GET("/", func(c *gin.Context) {
		serveEmbeddedWebIndex(c)
	})
	router.HEAD("/", func(c *gin.Context) {
		serveEmbeddedWebIndex(c)
	})
	router.GET("/favicon.ico", gin.WrapH(fileServer))
	router.GET("/assets/*filepath", gin.WrapH(fileServer))

	router.NoRoute(func(c *gin.Context) {
		if !isEmbeddedWebFallbackRequest(c) {
			c.Status(http.StatusNotFound)
			return
		}
		serveEmbeddedWebIndex(c)
	})
}

func serveEmbeddedWebIndex(c *gin.Context) {
	index, err := fs.ReadFile(embeddedWebFS, embeddedWebIndex)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", index)
}

func isEmbeddedWebFallbackRequest(c *gin.Context) bool {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return false
	}
	method := c.Request.Method
	if method != http.MethodGet && method != http.MethodHead {
		return false
	}

	path := c.Request.URL.Path
	if path == "" {
		path = "/"
	}
	if path == "/api" || strings.HasPrefix(path, "/api/") {
		return false
	}
	if path == "/mcp" || strings.HasPrefix(path, "/mcp/") {
		return false
	}
	if path == "/favicon.ico" || path == "/assets" || strings.HasPrefix(path, webAssetsPathPrefix) {
		return false
	}

	return true
}
