package consul

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAuthMethod(t *testing.T) {
	var a AuthMethod
	s := a.String()
	assert.Equal(t, "TOKEN", s)

	b, err := a.MarshalText()
	require.Nil(t, err)
	assert.Equal(t, []byte("TOKEN"), b)

	assert.False(t, a.isRefreshable())

	err = a.UnmarshalText([]byte("KUBERNETES"))
	require.Nil(t, err)
	assert.True(t, a.isRefreshable())
}
