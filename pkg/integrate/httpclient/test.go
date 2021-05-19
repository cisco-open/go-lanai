package httpclient

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

func kitFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	uri := &url.URL{
		Scheme: "http",
		Host: instance,
		Path: "/europa/api/ping",
	}

	cl := httptransport.NewClient(
		"GET",
		uri,
		httptransport.EncodeJSONRequest,
		testRespDecodeFunc,
	)

	return cl.Endpoint(), nil, nil
}

func factory(inst *discovery.Instance) (endpoint.Endpoint, error) {
	ctxPath := ""
	if inst.Meta != nil {
		ctxPath = inst.Meta[discovery.InstanceMetaKeyContextPath]
	}

	// TODO choose http or https based on tag "secure"
	uri := &url.URL{
		Scheme: "http",
		Host: fmt.Sprintf("%s:%d", inst.Address, inst.Port),
		Path: path.Clean(fmt.Sprintf("%s%s", ctxPath, "/api/ping")),
	}

	cl := httptransport.NewClient(
		"GET",
		uri,
		httptransport.EncodeJSONRequest,
		testRespDecodeFunc,
	)

	return cl.Endpoint(), nil
}


type TestHttpClient struct {
	serviceName string
	discClient discovery.Client
	consulClient kitconsul.Client
}

func NewTestHttpClient(client discovery.Client, conn *consul.Connection) *TestHttpClient {
	return &TestHttpClient{
		serviceName: "europa",
		discClient: client,
		consulClient: kitconsul.NewClient(conn.Client()),
	}
}

func (c *TestHttpClient) DoTest(ctx context.Context) error {
	//instancer := kitconsul.NewInstancer(c.consulClient, logger, "europa", []string{}, false)
	//endpointer := sd.NewEndpointer(instancer, kitFactory, logger, sd.InvalidateOnError(10 * time.Second))

	instancer, e := c.discClient.Instancer("europa")
	if e != nil {
		return e
	}

	endpointer, e := NewKitEndpointer(instancer, func(opts *EndpointerOption) {
		opts.ServiceName = c.serviceName
		opts.EndpointFactory = factory
		opts.Selector = discovery.InstanceIsHealthy()
		opts.InvalidateOnError = true
		opts.Logger = logger
	})
	if e != nil {
		return e
	}

	b := lb.NewRoundRobin(endpointer)
	//return lb.Retry(3, 1 * time.Minute, b)
	ep, e := b.Endpoint()
	if e != nil {
		return e
	} else if ep == nil {
		return nil
	}

	req := struct{}{}
	resp, e := ep(ctx, req)
	logger.Infof("Response: %v, Err: %v", resp, e)
	return nil
}

func testRespDecodeFunc(ctx context.Context, resp *http.Response) (response interface{}, err error) {
	defer func() {
		if e := resp.Body.Close(); e != nil {
			err = e
		}
	}()

	data, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, e
	}

	return string(data), nil
}

