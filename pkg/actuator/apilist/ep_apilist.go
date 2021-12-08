package apilist

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
)

const (
	ID                   = "apilist"
	EnableByDefault      = false
)

// ApiListEndpoint implements actuator.Endpoint, actuator.WebEndpoint
//goland:noinspection GoNameStartsWithPackageName
type ApiListEndpoint struct {
	actuator.WebEndpointBase
	staticPath string
}

func newEndpoint(di regDI) *ApiListEndpoint {
	if !fs.ValidPath(di.Properties.StaticPath) {
		panic("invalid static-path for apilist endpoint")
	}
	ep := ApiListEndpoint{
		staticPath: di.Properties.StaticPath,
	}
	ep.WebEndpointBase = actuator.MakeWebEndpointBase(func(opt *actuator.EndpointOption) {
		opt.Id = ID
		opt.Ops = []actuator.Operation{
			actuator.NewReadOperation(ep.Read),
		}
		opt.Properties = &di.MgtProperties.Endpoints
		opt.EnabledByDefault = EnableByDefault
	})
	return &ep
}

// Read never returns error
func (ep *ApiListEndpoint) Read(ctx context.Context, _ *struct{}) (interface{}, error) {
	resp, e := parseFromStaticFile(ep.staticPath)
	if e != nil {
		// Note we don't expose error. Instead, we return 404 like nothing is there
		logger.WithContext(ctx).Warnf(`unable to load static API list file "%s": %v`, ep.staticPath, e)
		return nil, web.NewHttpError(http.StatusNotFound, fmt.Errorf("APIList is not available"))
	}
	return resp, nil
}

func parseFromStaticFile(path string) (ret interface{}, err error) {
	// open
	var file fs.File
	var e error
	for _, fsys := range staticFS {
		if file, e = fsys.Open(path); e == nil {
			break
		}
	}
	if e != nil {
		return nil, e
	}

	// read
	defer func(){ _ = file.Close() }()
	decoder := json.NewDecoder(file)
	if e := decoder.Decode(&ret); e != nil {
		return nil, e
	}
	return
}

