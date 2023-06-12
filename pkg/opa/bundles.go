package opa

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"embed"
	"fmt"
	sdktest "github.com/open-policy-agent/opa/sdk/test"
	"go.uber.org/fx"
	"io/fs"
	"path/filepath"
	"strings"
)

//TODO this is just a POC, bundles should be loaded from bundle server

//go:embed bundle/.manifest bundle/roles bundle/operations bundle/tenancy  bundle/ownership bundle/api.rev.2 bundle/poc
var BundleFS embed.FS

var Bundles = map[string]embed.FS {
	"/bundles/bundle.tar.gz": BundleFS,
}

type BundleServerOut struct {
	fx.Out
	Server *sdktest.Server
}

func ProvideBundleServer(appCtx *bootstrap.ApplicationContext) (BundleServerOut, error){
	opts := make([]func(*sdktest.Server) error, 0, 5)
	for name, fsys := range Bundles {
		policies, e := loadBundleFiles(fsys)
		if e != nil {
			logger.WithContext(appCtx).Warnf("unable to load bundle [%s]: ", name, e)
			continue
		}
		opts = append(opts, sdktest.MockBundle(name, policies))
	}
	if len(opts) == 0 {
		return BundleServerOut{}, fmt.Errorf("failed to start OPA bundle server, unable to load any bundle")
	}

	ready := make(chan struct{}, 1)
	defer func() {close(ready)}()
	opts = append(opts, sdktest.Ready(ready))
	server, e := sdktest.NewServer(opts...)
	if e != nil {
		return BundleServerOut{}, fmt.Errorf("failed to start OPA bundle server: %v", e)
	}
	logger.WithContext(appCtx).Infof("OPA Bundles served at %q", server.URL())
	return BundleServerOut{
		Server: server,
	}, nil
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
