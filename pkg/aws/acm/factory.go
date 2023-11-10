package acm

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/sts"
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
			role = os.Getenv("AWS_ROLE_ARN")
		}
		token, err := os.ReadFile(path)
		if err != nil {
			logger.WithContext(ctx).Errorf("Failed to read web identity token file: %s", err.Error())
			return nil, err
		}
		sess, err := session.NewSession(cfg)
		if err != nil {
			return nil, err
		}
		params := &sts.AssumeRoleWithWebIdentityInput{
			RoleArn:          aws.String(role),
			RoleSessionName:  aws.String(f.properties.Credentials.RoleSessionName),
			WebIdentityToken: aws.String(string(token)),
		}
		svc := sts.New(sess)
		assumedRole, err := svc.AssumeRoleWithWebIdentity(params)
		if err != nil {
			logger.WithContext(ctx).Errorf("Failed to assume role via STS: %s", err.Error())
			return nil, err
		}
		cred = credentials.NewStaticCredentials(*assumedRole.Credentials.AccessKeyId, *assumedRole.Credentials.SecretAccessKey, *assumedRole.Credentials.SessionToken)
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
