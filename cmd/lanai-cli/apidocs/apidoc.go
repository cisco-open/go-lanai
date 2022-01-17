package apidocs

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	cache = map[string]*apidoc{}
)

type apidoc struct {
	source string
	value  map[string]interface{}
}

func writeApiDocLocal(_ context.Context, doc *apidoc) (string, error) {
	// create file or open and truncate
	absPath, file, e := cmdutils.OpenFile(doc.source, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if e != nil {
		return "", fmt.Errorf("unable to write API doc to file [%s]: %v", doc.source, e)
	}
	defer func() { _ = file.Close() }()

	switch fileExt := strings.ToLower(path.Ext(absPath)); fileExt {
	case ".yml", ".yaml":
		var data []byte
		data, e = yaml.Marshal(doc.value)
		if e == nil {
			_, e = file.Write(data)
		}
	case ".json", ".json5":
		e = json.NewEncoder(file).Encode(doc.value)
	default:
		return "", fmt.Errorf("unsupported file extension for OAS document: %s", fileExt)
	}
	if e != nil {
		return "", fmt.Errorf("cannot save document to [%s]: %v", absPath, e)
	}
	return absPath, nil
}

func loadApiDocs(ctx context.Context, paths []string) ([]*apidoc, error) {
	docs := make([]*apidoc, len(paths), len(paths) * 2)
	for i, p := range paths {
		doc, e := loadApiDoc(ctx, p)
		if e != nil {
			return nil, e
		}
		docs[i] = doc
		cache[doc.source] = doc
	}
	return docs, nil
}

func loadApiDoc(ctx context.Context, path string) (doc *apidoc, err error) {
	switch {
	case strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://"):
	// TODO load from remote
	case strings.HasPrefix(path, "git@"):
	// TODO load from git
	default:
		return loadApiDocLocal(ctx, path)
	}
	return
}

func loadApiDocLocal(_ context.Context, fPath string) (*apidoc, error) {
	absPath, e := filepath.Abs(path.Join(cmdutils.GlobalArgs.WorkingDir, fPath))
	if e != nil {
		return nil, fmt.Errorf("unable to resolve absolute path of file [%s]: %v", fPath, e)
	}
	if cached, ok := cache[absPath]; ok && cached != nil {
		return cached, nil
	}

	doc := apidoc{
		source: absPath,
	}
	switch fileExt := strings.ToLower(path.Ext(fPath)); fileExt {
	case ".yml", ".yaml":
		_, e = cmdutils.BindYamlFile(&doc.value, fPath)
	case ".json", ".json5":
		_, e = cmdutils.BindJsonFile(&doc.value, fPath)
	default:
		return nil, fmt.Errorf("unsupported file extension for OAS document: %s", fileExt)
	}
	if e != nil {
		return nil, e
	}
	return &doc, nil
}
