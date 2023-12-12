package acm

import (
	"context"
	awsclient "cto-github.cisco.com/NFV-BU/go-lanai/pkg/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
)

type ClientFactory interface {
	New(ctx context.Context, opts...func(opt *acm.Options)) (*acm.Client, error)
}

func NewClientFactory(loader awsclient.ConfigLoader) ClientFactory {
	return &acmFactory{
		configLoader: loader,
	}
}

type acmFactory struct {
	configLoader awsclient.ConfigLoader
}

func (f *acmFactory) New(ctx context.Context, opts...func(opt *acm.Options)) (*acm.Client, error) {
	cfg, e := f.configLoader.Load(ctx)
	if e != nil {
		return nil, e
	}
	return acm.NewFromConfig(cfg, opts...), nil
}
