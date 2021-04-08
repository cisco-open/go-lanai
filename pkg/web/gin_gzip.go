package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GzipAsset() gin.HandlerFunc {
	return func(gc *gin.Context) {
		if !isGzipAsset(gc.Request) {
			return
		}
		gc.Header("Content-Encoding", "gzip")
		gc.Header("Vary", "Accept-Encoding")
	}
}

func isGzipAsset(req *http.Request) bool {
	if !strings.HasSuffix(req.URL.Path, ".gz") {
		return false
	}

	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(req.Header.Get("Content-Type"), "text/event-stream") {

		return false
	}

	return true
}
