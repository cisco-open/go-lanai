package ittest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	secit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"go.uber.org/fx"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

type RecordingHttpClientCustomizer struct {
	Recorder *recorder.Recorder
}

func (c RecordingHttpClientCustomizer) Customize(opt *httpclient.ClientOption) {
	opt.HTTPClient = c.Recorder.GetDefaultClient()
}

func WithRecordedScopes() test.Options {
	fxOpts := []fx.Option{
		fx.Provide(jwt.BindCryptoProperties),
		fx.Provide(provideScopeDI),
	}

	opts := []test.Options{
		apptest.WithModules(scope.Module, seclient.Module),
		apptest.WithFxOptions(fxOpts...),
	}
	return func(opt *test.T) {
		for _, fn := range opts {
			fn(opt)
		}
	}
}

/*************************
	Providers
 *************************/

type scopeDI struct {
	fx.In
	ItProperties     secit.SecurityIntegrationProperties
	CryptoProperties jwt.CryptoProperties `optional:"true"`
	HttpClient httpclient.Client
	Recorder *recorder.Recorder `optional:"true"`
}

type scopeDIOut struct {
	fx.Out
	TokenReader oauth2.TokenStoreReader
	JwkStore    jwt.JwkStore
}

func provideScopeDI(di scopeDI) scopeDIOut {
	jwkStore := jwt.NewFileJwkStore(di.CryptoProperties)
	return scopeDIOut{
		JwkStore: jwkStore,
		TokenReader: NewRemoteTokenStoreReader(func(opt *RemoteTokenStoreOption) {
			opt.SkipRemoteCheck = true
			opt.HttpClient = di.HttpClient
			opt.BaseUrl = di.ItProperties.Endpoints.BaseUrl
			opt.ServiceName = di.ItProperties.ServiceName
			opt.ClientId = di.ItProperties.Client.ClientId
			opt.ClientSecret = di.ItProperties.Client.ClientSecret
			if di.Recorder != nil {
				opt.HttpClientConfig = &httpclient.ClientConfig{
					HTTPClient:  di.Recorder.GetDefaultClient(),
				}
			}
		}),
	}
}
