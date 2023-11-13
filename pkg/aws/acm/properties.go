package acm

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	ConfigRootACM = "aws.acm"
)

// AwsProperties describes common config used to consume AWS services
type AcmProperties struct {
	//Region for AWS client defaults to us-east-1
	Region string `json:"region"`
	//Endpoint for AWS client default empty can be used to override if consuming localstack
	Endpoint string `json:"endpoint"`
	//Credentials to be used to authenticate the AWS client
	Credentials Credentials `json:"credentials"`
}

// Credentials defines the type of credentials to use for AWS
type Credentials struct {
	//Type is one of static, env or sts.  Defaults to env.
	Type string `json:"type"`

	//The following is only relevant to static credential
	//Id is the AWS_ACCESS_KEY_ID for the account
	Id string `json:"id"`
	//Secret is the AWS_SECRET_ACCESS_KEY
	Secret string `json:"secret"`

	//The follow is relevant to sts credentials (Used in EKS)
	//RoleARN defines role to be assumed by application if omitted environment variable AWS_ROLE_ARN will be used
	RoleARN string `json:"role-arn"`
	//TokenFile is the path to the STS OIDC token file if omitted environment variable AWS_WEB_IDENTITY_TOKEN_FILE will be used
	TokenFile string `json:"token-file"`
	//RoleSessionName username to associate with session e.g. service account
	RoleSessionName string `json:"role-session-name"`
}

func NewAwsProperties(ctx *bootstrap.ApplicationContext) AcmProperties {
	return AcmProperties{
		Region: "us-east-1",
		Credentials: Credentials{
			Type: "env",
		},
	}
}

func BindAwsProperties(ctx *bootstrap.ApplicationContext) AcmProperties {
	props := NewAwsProperties(ctx)
	if err := ctx.Config().Bind(&props, ConfigRootACM); err != nil {
		panic(errors.Wrap(err, "failed to bind aws.AwsProperties"))
	}
	return props
}
