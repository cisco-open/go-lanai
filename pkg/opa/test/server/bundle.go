package opatestserver

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
	sdktest "github.com/open-policy-agent/opa/sdk/test"
	"go.uber.org/fx"
	"io/fs"
	"path/filepath"
	"strings"
)

var logger = log.New("OPA.Test")

type BundleServerOptions func(cfg *BundleServerConfig)
type BundleServerConfig struct {
	Bundles map[string]fs.FS
}

// WithBundleFS is a BundleServerOptions to add bundles from system
func WithBundleFS(name string, fsys fs.FS) BundleServerOptions {
	return func(cfg *BundleServerConfig) {
		cfg.Bundles[name] = fsys
	}
}

func NewBundleServer(ctx context.Context, opts ...BundleServerOptions) (*sdktest.Server, error) {
	cfg := BundleServerConfig{
		Bundles: map[string]fs.FS{},
	}
	for _, fn := range opts {
		fn(&cfg)
	}

	svrOpts := make([]func(*sdktest.Server) error, 0, 5)
	for name, fsys := range cfg.Bundles {
		policies, e := loadBundleFiles(fsys)
		if e != nil {
			logger.WithContext(ctx).Warnf("unable to load bundle [%s]: ", name, e)
			continue
		}
		svrOpts = append(svrOpts, sdktest.MockBundle(name, policies))
	}
	if len(svrOpts) == 0 {
		return nil, fmt.Errorf("failed to start OPA bundle server, unable to load any bundle")
	}

	ready := make(chan struct{}, 1)
	defer func() { close(ready) }()
	svrOpts = append(svrOpts, sdktest.Ready(ready))
	server, e := sdktest.NewServer(svrOpts...)
	if e != nil {
		return nil, fmt.Errorf("failed to start OPA bundle server: %v", e)
	}
	logger.WithContext(ctx).Infof("OPA Bundles served at %q", server.URL())
	return server, nil
}

func InitializeBundleServer(lc fx.Lifecycle, server *sdktest.Server) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			server.Stop()
			return nil
		},
	})
}

func loadBundleFiles(fsys fs.FS) (map[string]string, error) {
	// find and read all files
	files := map[string][]byte{}
	rootPath := "."
	e := fs.WalkDir(fsys, rootPath, func(path string, d fs.DirEntry, _ error) error {
		// Note we ignore any error and let it walk through entire tree
		if d.IsDir() {
			return nil
		}
		data, e := fs.ReadFile(fsys, path)
		if e != nil {
			return nil
		}
		if d.Name() == ".manifest" {
			rootPath = filepath.Dir(path)
		}
		files[path] = data
		return nil
	})
	if e != nil {
		return nil, e
	} else if len(files) == 0 {
		return nil, fmt.Errorf("no files was found in bundle FS")
	}

	// prepare bundle content
	ret := map[string]string{}
	for path, data := range files {
		name, e := filepath.Rel(rootPath, path)
		if e != nil {
			name = path
		}
		if strings.HasSuffix(name, ".json") {
			// nested data documents are not implemented in the dummy server
			name = strings.ReplaceAll(path, "/", "_")
		}
		ret[name] = string(data)
	}
	return ret, nil
}
