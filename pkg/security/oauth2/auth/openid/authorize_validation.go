package openid

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// OpenIDAuthorizeRequestProcessor implements AuthorizeRequestProcessor and order.Ordered
// it validate auth request against standard oauth2 specs
//goland:noinspection GoNameStartsWithPackageName
type OpenIDAuthorizeRequestProcessor struct {
	issuer security.Issuer
}

type ARPOptions func(*ARPOption)

type ARPOption struct {
	Issuer security.Issuer
}

func NewOpenIDAuthorizeRequestProcessor(opts...ARPOptions) *OpenIDAuthorizeRequestProcessor {
	opt := ARPOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &OpenIDAuthorizeRequestProcessor{
		issuer: opt.Issuer,
	}
}

func (p *OpenIDAuthorizeRequestProcessor) Process(ctx context.Context, request *auth.AuthorizeRequest) (validated *auth.AuthorizeRequest, err error) {
	if e := p.validateResponseTypes(ctx, request); e != nil {
		return nil, e
	}

	if e := p.validateImplicitFlow(ctx, request); e != nil {
		return nil, e
	}

	if e := p.validateDisplay(ctx, request); e != nil {
		return nil, e
	}

	if e := p.validateClaims(ctx, request); e != nil {
		return nil, e
	}

	if e := p.validateAcrValues(ctx, request); e != nil {
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

func (p *OpenIDAuthorizeRequestProcessor) validateResponseTypes(_ context.Context, request *auth.AuthorizeRequest) error {
	if request.ResponseTypes == nil {
		return oauth2.NewInvalidAuthorizeRequestError("response_type is required")
	}

	for k := range request.ResponseTypes {
		if !ResponseTypes.Has(strings.ToLower(k)) {
			return oauth2.NewInvalidResponseTypeError(fmt.Sprintf("unsupported response type: %s", k))
		}
	}
	return nil
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
func (p *OpenIDAuthorizeRequestProcessor) validateClaims(_ context.Context, request *auth.AuthorizeRequest) error {
	raw, ok := request.Parameters[oauth2.ParameterClaims]
	if !ok {
		return nil
	}

	cr := claimsRequest{}
	if e := json.Unmarshal([]byte(raw), &cr); e != nil {
		// maybe we should ignore this error
		return oauth2.NewInvalidAuthorizeRequestError(`invalid "claims" parameter`)
	}

	// set as extension
	if len(cr.UserInfo) != 0 {
		request.Extensions[oauth2.ExtRequestedUserInfoClaims] = cr.UserInfo
	}

	if len(cr.IdToken) != 0 {
		request.Extensions[oauth2.ExtRequestedIdTokenClaims] = cr.IdToken
	}
	return nil
}

func (p *OpenIDAuthorizeRequestProcessor) validateAcrValues(_ context.Context, request *auth.AuthorizeRequest) error {
	acrVals, ok := request.Parameters[oauth2.ParameterACR]
	if !ok {
		return nil
	}

	required := utils.NewStringSet()
	optional := utils.NewStringSet(strings.Split(acrVals, " ")...)
	optional.Remove("")
	if rc, ok := request.Extensions[oauth2.ExtRequestedIdTokenClaims].(claims.RequestedClaims); !ok {
		if acr, ok := rc.Get(oauth2.ClaimAuthCtxClassRef); ok && !acr.IsDefault() {
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
	if !possible {
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
		return nil
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
	if prompts.Has(PromptLogin) && isCurrentlyAuthenticated(ctx) {
		security.Clear(ctx)
		// TODO break the "login loop"
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
