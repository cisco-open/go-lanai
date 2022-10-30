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
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"net/http"
	"strings"
	"time"
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
		fx.Provide(provideScopeVCROptions),
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
	HttpClient       httpclient.Client
	Recorder         *recorder.Recorder `optional:"true"`
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
					HTTPClient: di.Recorder.GetDefaultClient(),
				}
			}
		}),
	}
}

type scopeVCROptionsOut struct {
	fx.Out
	VCROptions HttpVCROptions `group:"http-vcr"`
}

func provideScopeVCROptions() scopeVCROptionsOut {
	return scopeVCROptionsOut{
		VCROptions: HttpRecorderHooks(NewRecorderHook(extendedTokenValidityHook(), recorder.BeforeResponseReplayHook)),
	}
}

/*************************
	Additional Hooks
 *************************/

// extendedTokenValidityHook HTTP VCR hook that extend token validity to a distant future.
// During scope switching, token's expiry time is used to determine if token need to be refreshed.
// This would cause inconsistent HTTP interactions between recording time and replay time (after token expires)
// "expiry" and "expires_in" are JSON fields in `/v2/token` response and `exp` is a standard claim in `/v2/check_token` response
func extendedTokenValidityHook() func(i *cassette.Interaction) error {
	longValidity := 100 * 24 * 365 * time.Hour
	expiry := time.Now().Add(longValidity)
	tokenBodySanitizers := map[string]ValueSanitizer{
		"expiry":     SubstituteValueSanitizer(expiry.Format(time.RFC3339)),
		"expires_in": SubstituteValueSanitizer(longValidity.Seconds()),
		"exp":        SubstituteValueSanitizer(expiry.Unix()),
	}
	tokenBodyJsonPaths := parseJsonPaths([]string{"$.expiry", "$.expires_in", "$.exp"})
	return func(i *cassette.Interaction) error {
		if i.Response.Code != http.StatusOK ||
			!strings.Contains(i.Request.URL, "/v2/token") && !strings.Contains(i.Request.URL, "/v2/check_token") {
			return nil
		}
		i.Response.Body = sanitizeJsonBody(i.Response.Body, tokenBodySanitizers, tokenBodyJsonPaths)
		return nil
	}
}
