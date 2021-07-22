package claims

import (
	"context"
	"crypto"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/base64"
	"fmt"
	"strings"
)

// AddressClaim is defined at https://openid.net/specs/openid-connect-core-1_0.html#AddressClaim
type AddressClaim struct {
	Formatted  string `json:"formatted,omitempty"`
	StreetAddr string `json:"street_address,omitempty"`
	City       string `json:"locality,omitempty"`
	Region     string `json:"region,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
}

func AuthenticationTime(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.AuthenticationDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.AuthenticationTime(), errorMissingDetails)
}

func Nonce(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.OAuth2Request() == nil || opt.Source.OAuth2Request().Parameters() == nil {
		return nil, errorMissingRequest
	}

	nonce, _ := opt.Source.OAuth2Request().Parameters()[oauth2.ParameterNonce]
	return nonZeroOrError(nonce, errorMissingRequestParams)
}

func AuthContextClassRef(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.OAuth2Request() == nil || opt.Source.OAuth2Request().Parameters() == nil {
		return nil, errorMissingRequest
	}

	nonce, _ := opt.Source.OAuth2Request().Parameters()[oauth2.ParameterNonce]
	return nonZeroOrError(nonce, errorMissingRequestParams)
}

func AccessTokenHash(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	token := extractAccessToken(opt)
	if token == nil || token.Value() == "" {
		return nil, errorMissingToken
	}

	return calculateAccessTokenHash(token.Value())
}

func FullName(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	name := strings.TrimSpace(strings.Join([]string{details.FirstName(), details.LastName()}, " "))
	return nonZeroOrError(name, errorMissingDetails)
}

func FirstName(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.FirstName(), errorMissingDetails)
}

func LastName(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.LastName(), errorMissingDetails)
}

func Email(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.Email(), errorMissingDetails)
}

func EmailVerified(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return utils.BoolPtr(strings.TrimSpace(details.Email()) != ""), nil
}

func ZoneInfo(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	// TODO maybe impelment this if possibile to extract it from locale
	return nil, errorMissingDetails
}

func Locale(_ context.Context, opt *FactoryOption) (v interface{}, err error) {
	details, ok := opt.Source.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.LocaleCode(), errorMissingDetails)
}

func Address(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	acct, ok := tryReloadAccount(ctx, opt).(security.AccountMetadata)
	if !ok || acct == nil {
		return nil, errorMissingDetails
	}
	addr := AddressClaim{
		Formatted:  acct.LocaleCode(),
		//StreetAddr: "",
		//City:       "",
		//Region:     "",
		//PostalCode: "",
		//Country:    "",
	}
	return &addr, nil
}

/********************
	Helpers
 ********************/

var (
	jwtHashAlgorithms = map[string]crypto.Hash {
		"RS256": crypto.SHA256,
		"ES256": crypto.SHA256,
		"HS256": crypto.SHA256,
		"PS256": crypto.SHA256,
		"RS384": crypto.SHA384,
		"HS384": crypto.SHA384,
		"RS512": crypto.SHA512,
		"HS512": crypto.SHA512,
	}

)

func calculateAccessTokenHash(token string) (string, error) {
	// find out hashing algorithm
	headers, e := jwt.ParseJwtHeaders(token)
	if e != nil {
		return "", e
	}
	tokenAlg, _ := headers["alg"].(string)
	alg, ok := jwtHashAlgorithms[tokenAlg]
	if !ok || !alg.Available() {
		return "", fmt.Errorf(`hash is unsupported for access token with alg="%s"`, tokenAlg)
	}

	// do hash and take the left half
	hash := alg.New()
	if _, e := hash.Write([]byte(token)); e != nil {
		return "", e
	}

	leftHalf := hash.Sum(nil)[hash.Size() / 2:]
	return base64.RawStdEncoding.EncodeToString(leftHalf), nil
}
