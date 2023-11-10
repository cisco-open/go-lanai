package consul

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
)

func TestKubernetesClient_Login(t *testing.T) {
	var fakeToken = "803c06c4-36d1-1946-c687-9199b4b7c256"
	var fakeAuthResp = `{"AccessorID":"5bc7b351-638d-379b-86ce-53379c61f1d1","SecretID":"` + fakeToken + `","Description":"token created via login","Roles":[{"ID":"8c1883da-708a-a230-42e9-f01d98a4c88e","Name":"example-role"}],"Local":true,"AuthMethod":"consul-k8s-component-auth-method","CreateTime":"2023-11-10T14:37:29.270353178Z","Hash":"h2FpCxPojSGRi0aFbLxWyoDeza5tcBQsUnUV7yDsLNk=","CreateIndex":410416,"ModifyIndex":410416}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fakeAuthResp))
	}))
	u, err := url.Parse(ts.URL)
	require.Nil(t, err)
	port, err := strconv.Atoi(u.Port())
	require.Nil(t, err)
	host := u.Hostname()
	f, err := os.CreateTemp("./", "testtoken")
	require.Nil(t, err)
	defer os.Remove(f.Name())
	testProps := &ConnectionProperties{Authentication: Kubernetes, Host: host, Scheme: u.Scheme, Port: port, Kubernetes: KubernetesConfig{JWTPath: f.Name()}}
	conn, err := NewConnection(testProps)
	ca := newClientAuthentication(testProps)
	token, err := ca.Login(conn.client)
	require.Nil(t, err)
	assert.Equal(t, fakeToken, token)
}
