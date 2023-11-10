package acm

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAwsSessionFactoryImpl_New(t *testing.T) {
	factory := NewAwsAcmFactory(
		AcmProperties{
			Region: "us-west-2",
			Credentials: Credentials{
				Type:   "static",
				Id:     "test",
				Secret: "test",
			},
		})

	ctx := context.Background()
	result, err := factory.New(ctx)

	require.NoError(t, err)
	assert.IsType(t, &acm.ACM{}, result)
	bctx := bootstrap.NewApplicationContext()
	acmclient := newDefaultClient(bctx, factory)
	assert.NotNil(t, acmclient)
}
