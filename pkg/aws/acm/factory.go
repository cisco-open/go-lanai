package acm

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"os"
)

type AwsAcmFactory interface {
	New(ctx context.Context) (*acm.ACM, error)
}

func NewAwsAcmFactory(p AcmProperties) AwsAcmFactory {
	return &awsSessionFactoryImpl{
		properties: p,
	}
}

type awsSessionFactoryImpl struct {
	properties AcmProperties
}

func (f *awsSessionFactoryImpl) New(ctx context.Context) (*acm.ACM, error) {
	cfg := aws.NewConfig()
	cfg.Region = aws.String(f.properties.Region)
	if f.properties.Endpoint != "" {
		cfg.Endpoint = aws.String(f.properties.Endpoint)
	}
	var cred *credentials.Credentials
	switch f.properties.Credentials.Type {
	case "static":
		cred = credentials.NewStaticCredentials(f.properties.Credentials.Id, f.properties.Credentials.Secret, "")

	case "sts":
		path := f.properties.Credentials.TokenFile
		if path == "" {
			path = os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
		}
		role := f.properties.Credentials.RoleARN
		if role == "" {
			role = os.Getenv("")
		}
	default:
		cred = credentials.NewEnvCredentials()
	}
	cfg.Credentials = cred

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}
	a := acm.New(sess)
	logger.WithContext(ctx).Info("New AWS ACM client created")
	return a, nil
}
