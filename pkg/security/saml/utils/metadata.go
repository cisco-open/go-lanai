package samlutils

import (
	"context"
	"github.com/crewjam/httperr"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type MetadataResolverOptions func(opt *MetadataResolverOption)
type MetadataResolverOption struct {
	HttpClient *http.Client
}

func WithHttpClient(client *http.Client) MetadataResolverOptions {
	return func(opt *MetadataResolverOption) {
		opt.HttpClient = client
	}
}

// ResolveMetadata try to resolve metadata from given metadata source
// Following modes are supported
// - if the source start with "<", it's treated as inline XML text
// - if the source is a valid HTTP/HTTPS URL, metadata is fetched over network using http.Client
// - if the source is a valid FILE URL (file://), metadata is loaded from file system
// - for any other source value, it's treated as file path
func ResolveMetadata(ctx context.Context, metadataSource string, opts...MetadataResolverOptions) (*saml.EntityDescriptor, []byte, error) {
	opt := MetadataResolverOption{
		HttpClient: http.DefaultClient,
	}
	for _, fn := range opts {
		fn(&opt)
	}
	if strings.HasPrefix(metadataSource, "<") {
		return ParseMetadataFromXml(metadataSource)
	}

	metadataUrl, err := url.Parse(metadataSource)
	if err != nil {
		return nil, nil, err
	}

	//if it's not url or file url, assume it's relative path
	if metadataUrl.Scheme == "file" || metadataUrl.Scheme == "" {
		return ParseMetadataFromFile(metadataUrl.Path)
	} else {
		return FetchMetadata(ctx, opt.HttpClient, metadataUrl)
	}
}

func ParseMetadataFromXml(xml string) (*saml.EntityDescriptor, []byte, error) {
	data := []byte(xml)
	metadata, err := samlsp.ParseMetadata(data)
	return metadata, data, err
}

func ParseMetadataFromFile(fileLocation string) (*saml.EntityDescriptor, []byte, error) {
	file, err := os.Open(fileLocation)
	if err != nil {
		return nil, nil, err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}
	metadata, err := samlsp.ParseMetadata(data)
	return metadata, data, err
}

func FetchMetadata(ctx context.Context, httpClient *http.Client, metadataURL *url.URL) (*saml.EntityDescriptor, []byte, error) {
	req, err := http.NewRequest("GET", metadataURL.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, nil, httperr.Response(*resp)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, data, err
	}

	metadata, err := samlsp.ParseMetadata(data)
	return metadata, data, err
}

