package consul

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTokenClientAuthentication_Login(t *testing.T) {
	var testtoken = "testtoken"
	testProps := &ConnectionProperties{Token: testtoken}
	ca := newClientAuthentication(testProps)
	token, err := ca.Login(nil)
	require.Nil(t, err)
	assert.Equal(t, testtoken, token)
}
