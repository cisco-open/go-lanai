package saml_util

import (
	"context"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func ParseMetadataFromFile(fileLocation string) (*saml.EntityDescriptor, error){
	file, err := os.Open(fileLocation)
	if err != nil {
		return nil, err
	}
	raw, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	metadata, err := samlsp.ParseMetadata(raw)
	return metadata, err
}

func ResolveMetadata(metadataSource string, httpClient *http.Client) (*saml.EntityDescriptor, error) {
	//TODO: support the case where the literal metadata source is pasted?
	metadataUrl, err := url.Parse(metadataSource)
	if err != nil {
		return nil, err
	}
	//if it's not url or file url, assume it's relative path
	if metadataUrl.Scheme == "file" || metadataUrl.Scheme == "" {
		return ParseMetadataFromFile(metadataUrl.Path)
	} else {
		metadata, err := samlsp.FetchMetadata(context.TODO(), httpClient, *metadataUrl)
		return metadata, err
	}
}