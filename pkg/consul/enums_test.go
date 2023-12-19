package consul

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAuthMethod(t *testing.T) {
	var a = Token
	s := fmt.Sprintf("%v", a)
	assert.Equal(t, "token", s)
	assert.False(t, a.isRefreshable())

	err := a.UnmarshalText([]byte("kubernetes"))
	require.Nil(t, err)
	assert.True(t, a.isRefreshable())
}
