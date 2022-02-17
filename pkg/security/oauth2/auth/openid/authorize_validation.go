package openid

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	keyPromptProcessed = "X-OIDC-PROMPT-PROCESSED"
)

var (
	supportedResponseTypes = utils.NewStringSet("id_token", "token", "code")
)

// OpenIDAuthorizeRequestProcessor implements ChainedAuthorizeRequestProcessor and order.Ordered
// it validate auth request against standard oauth2 specs
//goland:noinspection GoNameStartsWithPackageName
type OpenIDAuthorizeRequestProcessor struct {
	issuer             security.Issuer
	jwtDecoder         jwt.JwtDecoder
	fallbackJwtDecoder jwt.JwtDecoder
}

type ARPOptions func(opt *ARPOption)

type ARPOption struct {
	Issuer     security.Issuer
	JwtDecoder jwt.JwtDecoder
}

func NewOpenIDAuthorizeRequestProcessor(opts ...ARPOptions) *OpenIDAuthorizeRequestProcessor {
	opt := ARPOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &OpenIDAuthorizeRequestProcessor{
		issuer:             opt.Issuer,
		jwtDecoder:         opt.JwtDecoder,
		fallbackJwtDecoder: jwt.NewPlaintextJwtDecoder(),
	}
}

func (p *OpenIDAuthorizeRequestProcessor) Process(ctx context.Context, request *auth.AuthorizeRequest, chain auth.AuthorizeRequestProcessChain) (validated *auth.AuthorizeRequest, err error) {
	// first thing first, is "openid" scope requested?
	if !request.Scopes.Has(oauth2.ScopeOidc) {
		return chain.Next(ctx, request)
	}

	if e := p.validateResponseTypes(ctx, request); e != nil {
		return nil, e
	}

	// attempt to decode from request object
	if request, err = p.decodeRequestObject(ctx, request); err != nil {
		return
	}

	// continue with the chain
	if request, err = chain.Next(ctx, request); err != nil {
		return
	}

	// additional checks
	if e := p.validateImplicitFlow(ctx, request); e != nil {
		return nil, e
	}

	if e := p.validateDisplay(ctx, request); e != nil {
		return nil, e
	}

	cr, e := p.validateClaims(ctx, request)
	if e != nil {
		return nil, e
	}

	if e := p.validateAcrValues(ctx, request, cr); e != nil {
		return nil, e
	}

	if e := p.processMaxAge(ctx, request); e != nil {
		return nil, e
	}

	if e := p.processPrompt(ctx, request); e != nil {
		return nil, e
	}
	return request, nil
}

func (p *OpenIDAuthorizeRequestProcessor) decodeRequestObject(ctx context.Context, request *auth.AuthorizeRequest) (*auth.AuthorizeRequest, error) {
	reqUri, uriOk := request.Parameters[oauth2.ParameterRequestUri]
	reqObj, objOk := request.Parameters[oauth2.ParameterRequestObj]
	switch {
	case !uriOk && !objOk:
		return request, nil
	case uriOk && objOk:
		return nil, oauth2.NewInvalidAuthorizeRequestError(fmt.Errorf("%s and %s are exclusive", oauth2.ParameterRequestUri, oauth2.ParameterRequestObj))
	case uriOk:
		if strings.HasPrefix(strings.ToLower(reqUri), "https:") {
			return nil, oauth2.NewInvalidAuthorizeRequestError(fmt.Errorf("%s must use https", oauth2.ParameterRequestUri))
		}
		bytes, e := httpGet(ctx, reqUri)
		if e != nil {
			return nil, oauth2.NewInvalidAuthorizeRequestError(fmt.Errorf("unable to fetch request object from %s: %v", oauth2.ParameterRequestUri, e))
		}
		reqObj = string(bytes)
	}

	// decode JWT using configured decoder, fallback to plaintext decoder
	claims, e := p.jwtDecoder.Decode(ctx, reqObj)
	if e != nil {
		if claims, e = p.fallbackJwtDecoder.Decode(ctx, reqObj); e != nil {
			return nil, oauth2.NewInvalidAuthorizeRequestError(fmt.Errorf("invalid request object: %v", e))
		}
	}

	//nolint:contextcheck
	decoded, e := claimsToAuthRequest(request.Context(), claims)
	if e != nil {
		return nil, oauth2.NewInvalidAuthorizeRequestError(fmt.Errorf("invalid request object: %v", e))
	}

	switch {
	case !request.ResponseTypes.Equals(decoded.ResponseTypes):
		return nil, oauth2.NewInvalidAuthorizeRequestError(fmt.Errorf("invalid request object - inconsistant response type"))
	case !decoded.Scopes.Has(oauth2.ScopeOidc):
		return nil, oauth2.NewInvalidAuthorizeRequestError(fmt.Errorf("invalid request object - missing 'openid' scope"))
	}
	return decoded, nil
}

func (p *OpenIDAuthorizeRequestProcessor) validateResponseTypes(ctx context.Context, request *auth.AuthorizeRequest) error {
	return auth.ValidateResponseTypes(ctx, request, supportedResponseTypes)
}

func (p *OpenIDAuthorizeRequestProcessor) validateImplicitFlow(_ context.Context, request *auth.AuthorizeRequest) error {
	if !request.ResponseTypes.Has("id_token") && !request.ResponseTypes.Has("token") {
		return nil
	}

	// use of nonce is required when implicit flow is used without response type "code"
	nonce, ok := request.Parameters[oauth2.ParameterNonce]
	if !request.ResponseTypes.Has("code") && (!ok || nonce == "") {
		return oauth2.NewInvalidAuthorizeRequestError("nonce is required for implicit flow")
	}
	return nil
}

func (p *OpenIDAuthorizeRequestProcessor) validateDisplay(ctx context.Context, request *auth.AuthorizeRequest) error {
	display, ok := request.Parameters[oauth2.ParameterDisplay]
	if ok && display != "" && !SupportedDisplayMode.Has(display) {
		logger.WithContext(ctx).Infof("unsupported display mode [%s] was requested.", display)
	}
	return nil
}

// https://openid.net/specs/openid-connect-core-1_0.html#ClaimsParameter
func (p *OpenIDAuthorizeRequestProcessor) validateClaims(_ context.Context, request *auth.AuthorizeRequest) (*ClaimsRequest, error) {
	raw, ok := request.Parameters[oauth2.ParameterClaims]
	if !ok {
		return nil, nil
	}

	cr := ClaimsRequest{}
	if e := json.Unmarshal([]byte(raw), &cr); e != nil {
		// maybe we should ignore this error
		return nil, oauth2.NewInvalidAuthorizeRequestError(`invalid "claims" parameter`)
	}

	return &cr, nil
}

func (p *OpenIDAuthorizeRequestProcessor) validateAcrValues(_ context.Context, request *auth.AuthorizeRequest, claimsReq *ClaimsRequest) error {
	acrVals, ok := request.Parameters[oauth2.ParameterACR]
	if !ok {
		return nil
	}

	required := utils.NewStringSet()
	optional := utils.NewStringSet(strings.Split(acrVals, " ")...)
	optional.Remove("")
	if claimsReq != nil && len(claimsReq.IdToken) != 0 {
		if acr, ok := claimsReq.IdToken.Get(oauth2.ClaimAuthCtxClassRef); ok && !acr.IsDefault() {
			if acr.Essential() {
				required.Add(acr.Values()...)
			} else {
				optional.Add(acr.Values()...)
			}
		}
	}

	// Note, for now we only validate if required ACRs are possible, this is consistent with Java impl.
	supported := utils.NewStringSet(
		p.issuer.LevelOfAssurance(0),
		p.issuer.LevelOfAssurance(1),
		p.issuer.LevelOfAssurance(2),
	)
	if isMFAPossible() {
		supported.Add(p.issuer.LevelOfAssurance(3))
	}

	// if any required ACR level is supported, we allow the request
	possible := false
	for lvl := range required {
		if supported.Has(lvl) {
			possible = true
			break
		}
	}
	if len(required) != 0 && !possible {
		return oauth2.NewGranterNotAvailableError("requested acr level is not possible")
	}
	return nil
}

func (p *OpenIDAuthorizeRequestProcessor) processMaxAge(ctx context.Context, request *auth.AuthorizeRequest) error {
	maxAgeStr, ok := request.Parameters[oauth2.ParameterMaxAge]
	if !ok {
		return nil
	}

	maxAge, e := time.ParseDuration(fmt.Sprintf("%ss", maxAgeStr))
	if e != nil {
		return nil //nolint:nilerr // per OpenID specs, "authroize" endpoint should simply ignore invalid request params
	}

	current := security.Get(ctx)
	authTime := security.DetermineAuthenticationTime(ctx, current)
	if !security.IsFullyAuthenticated(current) || authTime.IsZero() {
		return nil
	}

	if authTime.Add(maxAge).Before(time.Now()) {
		security.Clear(ctx)
	}
	return nil
}

func (p *OpenIDAuthorizeRequestProcessor) processPrompt(ctx context.Context, request *auth.AuthorizeRequest) error {
	prompt, ok := request.Parameters[oauth2.ParameterPrompt]
	if !ok || prompt == "" {
		return nil
	}
	prompts := utils.NewStringSet(strings.Split(prompt, " ")...)

	// handle "none"
	if prompts.Has(PromptNone) && (len(prompts) > 1 || !isCurrentlyAuthenticated(ctx)) {
		return NewInteractionRequiredError("unable to authenticate without interact with user")
	}

	// handle "login"
	// to break the login loop, we put a special header to current http request and it will be saved by request cache
	if prompts.Has(PromptLogin) && !isPromptLoginProcessed(ctx) && isCurrentlyAuthenticated(ctx) {
		security.Clear(ctx)
		if e := setPromptLoginProcessed(ctx); e != nil {
			return NewLoginRequiredError("unable to initiate login")
		}
	}
	// We don't support "select_account" and "consent" is checked when we have decided to show user approval
	return nil
}

/*********************
	Helpers
 *********************/

func isCurrentlyAuthenticated(ctx context.Context) bool {
	return security.IsFullyAuthenticated(security.Get(ctx))
}

func isMFAPossible() bool {
	return true
}

func isPromptLoginProcessed(ctx context.Context) bool {
	req := getHttpRequest(ctx)
	if req == nil {
		return false
	}
	return PromptLogin == req.Header.Get(keyPromptProcessed)
}

func setPromptLoginProcessed(ctx context.Context) error {
	req := getHttpRequest(ctx)
	if req == nil {
		return fmt.Errorf("unable to extract http request")
	}
	req.Header.Set(keyPromptProcessed, PromptLogin)
	return nil
}

func getHttpRequest(ctx context.Context) *http.Request {
	if gc := web.GinContext(ctx); gc != nil {
		return gc.Request
	}
	return nil
}

func httpGet(ctx context.Context, urlStr string) ([]byte, error) {
	parsed, e := url.Parse(urlStr)
	if e != nil {
		return nil, e
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), http.NoBody)
	if e != nil {
		return nil, e
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return nil, e
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("non-2XX status code")
	}

	return ioutil.ReadAll(resp.Body)
}

func claimsToAuthRequest(ctx context.Context, claims oauth2.Claims) (*auth.AuthorizeRequest, error) {
	return auth.ParseAuthorizeRequestWithKVs(ctx, claims.Values())
}
