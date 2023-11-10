package acm

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAwsProperties(t *testing.T) {
	ctx := bootstrap.NewApplicationContext()
	props := NewAwsProperties(ctx)
	assert.Equal(t, "us-east-1", props.Region)
}
