package apidocs

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	schemaPrefixGitHub = "github://"
	schemaPrefixHttp   = "http://"
	schemaPrefixHttps  = "https://"
	kDefaultGitHubPAT  = "default"
)

const (
	extYaml    = ".yaml"
	extYamlAlt = ".yml"
	extJson    = ".json"
	extJson5   = ".json5"
)

var (
	cache          = map[string]*apidoc{}
	githubPatCache map[string]string
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
	case extYaml, extYamlAlt:
		var data []byte
		data, e = yaml.Marshal(doc.value)
		if e == nil {
			_, e = file.Write(data)
		}
	case extJson, extJson5:
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
	docs := make([]*apidoc, len(paths), len(paths)*2)
	for i, p := range paths {
		doc, e := loadApiDoc(ctx, p)
		if e != nil {
			return nil, e
		}
		docs[i] = doc
	}
	return docs, nil
}

func loadApiDoc(ctx context.Context, path string) (doc *apidoc, err error) {
	defer func() {
		if err == nil && doc != nil {
			cache[doc.source] = doc
		}
	}()
	switch {
	case strings.HasPrefix(path, schemaPrefixHttp) || strings.HasPrefix(path, schemaPrefixHttps):
		return loadApiDocHttp(ctx, path)
	case strings.HasPrefix(path, schemaPrefixGitHub):
		return loadApiDocGitHub(ctx, path)
	default:
		return loadApiDocLocal(ctx, path)
	}
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
	case extYaml, extYamlAlt:
		_, e = cmdutils.BindYamlFile(&doc.value, fPath)
	case extJson, extJson5:
		_, e = cmdutils.BindJsonFile(&doc.value, fPath)
	default:
		return nil, fmt.Errorf("unsupported file extension for OAS document: %s", fileExt)
	}
	if e != nil {
		return nil, e
	}
	return &doc, nil
}

func loadApiDocGitHub(ctx context.Context, rawUrl string) (*apidoc, error) {
	rawUrl = strings.Replace(rawUrl, schemaPrefixGitHub, schemaPrefixHttps, 1)
	return loadApiDocHttp(ctx, rawUrl, func(r *http.Request) {
		token := githubAccessToken(ctx, r.URL.Host)
		if len(token) != 0 {
			r.Header.Set("Authorization", fmt.Sprintf("token %s", token))
		}
	})
}

func loadApiDocHttp(ctx context.Context, rawUrl string, opts ...func(r *http.Request)) (*apidoc, error) {
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, rawUrl, nil)
	if e != nil {
		return nil, fmt.Errorf("invalid URL [%s]: %v", rawUrl, e)
	}

	urlStr := req.URL.String()
	if cached, ok := cache[urlStr]; ok && cached != nil {
		return cached, nil
	}

	// apply options
	for _, fn := range opts {
		fn(req)
	}

	// send request and check result
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return nil, fmt.Errorf("unable to GET requested URL [%s]: %v", urlStr, e)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("unable to Get requested URL [%s]: Status %d", urlStr, resp.StatusCode)
	}

	// parse format
	doc := apidoc{
		source: urlStr,
	}
	switch ext := contentTypeAsExt(req, resp); ext {
	case extYaml, extYamlAlt:
		e = cmdutils.BindYaml(resp.Body, &doc.value)
	case extJson, extJson5:
		e = json.NewDecoder(resp.Body).Decode(&doc.value)
	default:
		return nil, fmt.Errorf("unsupported file format for OAS document: %s", ext)
	}
	if e != nil {
		return nil, e
	}
	return &doc, nil
}

func contentTypeAsExt(req *http.Request, resp *http.Response) string {
	ct := resp.Header.Get("Content-Type")
	fileExt := strings.ToLower(path.Ext(req.URL.EscapedPath()))
	if mt, _, e := mime.ParseMediaType(ct); e == nil {
		switch {
		case mt == "application/json":
			fileExt = extJson
		case strings.HasSuffix(mt, "yaml") || strings.HasSuffix(mt, "yml"):
			fileExt = extYamlAlt
		}
	}
	return fileExt
}

func githubAccessToken(ctx context.Context, host string) string {
	// populate inmemory cache if possible
	if githubPatCache == nil {
		if e := populateGithubPatCache(); e != nil {
			logger.WithContext(ctx).Warnf("invalid GitHub access token: %v", e)
			return ""
		}
	}
	token, ok := githubPatCache[host]
	if !ok {
		return githubPatCache[kDefaultGitHubPAT]
	}
	return token
}

func populateGithubPatCache() error {
	githubPatCache = make(map[string]string)
	// parse from ResolveConfig
	for _, v := range ResolveConf.GitHubTokens {
		token := os.ExpandEnv(v.Token)
		if len(token) == 0 {
			continue
		}
		host := strings.ToLower(v.Host)
		if len(host) == 0 {
			host = kDefaultGitHubPAT
		}
		githubPatCache[host] = token
	}

	// parse from ResolveArguments
	for _, arg := range ResolveArgs.GitHubPATs {
		split := strings.SplitN(arg, "@", 2)
		val := os.ExpandEnv(split[0])
		if len(val) == 0 {
			continue
		}
		if len(split) == 1 {
			githubPatCache[kDefaultGitHubPAT] = val
		} else {
			githubPatCache[split[1]] = val
		}
	}
	return nil
}
