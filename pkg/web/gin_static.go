package web

import (
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
	"strings"
)

var (
	predefinedAliases = map[string]string {
		"index": "index.html",
	}
	predefinedExtensions = []string{
		".gz",
	}
)

type ginStaticAssetsHandler struct {
	rewriter RequestRewriter
	fsys     fs.FS
	aliases  map[string]string
}

func (h ginStaticAssetsHandler) FilenameRewriteHandlerFunc() gin.HandlerFunc {
	//prefix := h.calculateStripPrefix(basePath, relativePath)
	return func(gc *gin.Context) {
		file := gc.Param("filepath")
		if h.canRead(file) {
			return
		}

		// try aliases
		if handled := h.tryAliases(gc, h.aliases, file); handled {
			return
		}

		if handled := h.tryAliases(gc, predefinedAliases, file); handled {
			return
		}

		// try extensions
		h.tryExtensions(gc, predefinedExtensions, file)
	}
}

func (h ginStaticAssetsHandler) PreCompressedGzipAsset() gin.HandlerFunc {
	return func(gc *gin.Context) {
		if !h.isGzipAsset(gc.Request) {
			return
		}
		gc.Header("Content-Encoding", "gzip")
		gc.Header("Vary", "Accept-Encoding")
	}
}

func (h ginStaticAssetsHandler) isGzipAsset(req *http.Request) bool {
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

func (h ginStaticAssetsHandler) canRead(filePath string) bool {
	f, e := h.fsys.Open(filePath)
	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()
	return e == nil
}

func (h ginStaticAssetsHandler) tryAliases(gc *gin.Context, aliases map[string]string, file string) bool {
	for k, v := range aliases {
		if !strings.HasSuffix(file, k) {
			continue
		}

		alias := h.replaceLast(file, k, v)
		// to avoid infinite loop or unnecessary rewrite,
		// we check if alias is same as the original file and if the alias file path exists
		if alias == file || !h.canRead(alias) {
			continue
		}

		_ = h.rewrite(gc, k, v)
		return true
	}

	return false
}

func (h ginStaticAssetsHandler) tryExtensions(gc *gin.Context, extensions []string, file string) bool {
	for _, v := range extensions {
		alias := file + v
		// to avoid infinite loop or unnecessary rewrite,
		// we check if alias is same as the original file and if the alias file path exists
		if alias == file || !h.canRead(alias) {
			continue
		}

		_ = h.rewrite(gc, "", v)
		return true
	}

	return false
}

func (h ginStaticAssetsHandler) rewrite(gc *gin.Context, value, rewrite string) error {
	// make a url copy
	u := *gc.Request.URL
	u.Path = h.replaceLast(u.Path, value, rewrite)

	// handle rewrite
	request := gc.Request
	request.URL = &u
	return h.rewriter.HandleRewrite(request)
}

func (h ginStaticAssetsHandler) replaceLast(s, substr, replacement string) string {
	if substr == "" {
		return s + replacement
	}

	i := strings.LastIndex(s, substr)
	if i < 0 {
		return s
	}
	return s[:i] + replacement + s[i+len(substr):]
}