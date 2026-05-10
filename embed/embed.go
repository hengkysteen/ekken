package embed

import (
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func ServeEmbedded(engine *gin.Engine, mode string) {
	if mode != "production" || !hasEmbeddedUI() {
		return
	}

	distFS, err := fs.Sub(getEmbeddedUI(), "dist")
	if err != nil {
		return
	}

	fileServer := http.FileServer(http.FS(distFS))

	engine.NoRoute(func(c *gin.Context) {
		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		
		// If root, serve index.html
		if path == "" {
			path = "index.html"
		}

		// Try to see if file exists in embedded FS
		_, err := fs.Stat(distFS, path)
		if err == nil {
			// File exists, serve it using standard file server
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// If not found and it's an asset or API, return 404
		if strings.HasPrefix(path, "assets/") || strings.HasPrefix(path, "api/") || filepath.Ext(path) != "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// Otherwise, it's a SPA route, serve index.html
		indexData, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
	})
}

func contentType(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return "text/html; charset=utf-8"
	}
	ct := mime.TypeByExtension(ext)
	if ct == "" {
		ct = "application/octet-stream"
	}
	return ct
}
