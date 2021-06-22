package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"embed"
	"github.com/pkg/errors"
	"time"
)

const (
	PropertiesPrefix = "integrate.security"
)

//go:embed defaults-integrate-security.yml
var DefaultConfigFS embed.FS

//goland:noinspection GoNameStartsWithPackageName
type SecurityIntegrationProperties struct {
	// How much time after a failed attempt, when re-try is allowed. Before this period pass,
	// integration framework will not re-attempt switching context to same combination of username and tenant name
	FailureBackOff utils.Duration `json:"failure-back-off"`

	// How much time that security context is guaranteed to be valid after requested.
	// when such validity cannot be guaranteed (e.g. this value is longer than token's validity),
	// we use FailureBackOff and re-request new token after `back-off` passes
	GuaranteedValidity utils.Duration `json:"guaranteed-validity"`

	ServiceName string                      `json:"service-name"`
	Endpoints   AuthEndpointsProperties     `json:"endpoints"`
	Client      ClientCredentialsProperties `json:"client"`
	Accounts    AccountsProperties          `json:"accounts"`
}

type ClientCredentialsProperties struct {
	ClientId     string `json:"client-id"`
	ClientSecret string `json:"secret"`
}

type AuthEndpointsProperties struct {
	// BaseUrl is used to override service discovery and load-balancing, it should kept empty in production
	BaseUrl       string `json:"base-url"`
	PasswordLogin string `json:"password-login"`
	SwitchContext string `json:"switch-context"`
}

type AccountsProperties struct {
	Default    AccountCredentialsProperties   `json:"default"`
	Additional []AccountCredentialsProperties `json:"additional"`
}

type AccountCredentialsProperties struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	SystemAccount bool   `json:"system-account"`
}

//NewSecurityIntegrationProperties create a DataProperties with default values
func NewSecurityIntegrationProperties() *SecurityIntegrationProperties {
	return &SecurityIntegrationProperties{
		FailureBackOff:     utils.Duration(300 * time.Second),
		GuaranteedValidity: utils.Duration(30 * time.Second),
		ServiceName:        "europa",
		Endpoints:          AuthEndpointsProperties{
			PasswordLogin: "/v2/token",
			SwitchContext: "/v2/token",
		},
		Client:             ClientCredentialsProperties{
			ClientId:     "nfv-service",
			ClientSecret: "nfv-service-secret",
		},
		Accounts:           AccountsProperties{
			Default:    AccountCredentialsProperties{
				Username:      "system",
				Password:      "system",
				SystemAccount: true,
			},
		},
	}
}

//BindSecurityIntegrationProperties create and bind SessionProperties, with a optional prefix
func BindSecurityIntegrationProperties(ctx *bootstrap.ApplicationContext) SecurityIntegrationProperties {
	props := NewSecurityIntegrationProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SecurityIntegrationProperties"))
	}
	return *props
}


