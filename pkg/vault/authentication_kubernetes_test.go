package vault

import (
	"bytes"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/httpvcr/cassette"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/httpvcr/recorder"
	"github.com/hashicorp/vault/api"
	"github.com/onsi/gomega"
	"io"
	"net/http"
	"testing"
)

func matchVaultResponse(r *http.Request, i cassette.Request) bool {
	if r.Body == nil {
		return cassette.DefaultMatcher(r, i)
	}
	var b bytes.Buffer
	if _, err := b.ReadFrom(r.Body); err != nil {
		return false
	}
	r.Body = io.NopCloser(&b)
	return cassette.DefaultMatcher(r, i) && (b.String() == "" || b.String() == i.Body)
}

func TestKubernetesClient_Login(t *testing.T) {
	type args struct {
		role string
	}
	tests := []struct {
		name         string
		args         args
		config       KubernetesConfig
		playbackFile string
		recorderMode recorder.Mode
		wantErr      bool
	}{
		{
			name:         "Login should return a client token if successful",
			recorderMode: recorder.ModeReplaying,
			config: KubernetesConfig{
				JWTPath: "testdata/authentication_kubernetes/successful_client_token",
				Role:    "devweb-app",
			},
			playbackFile: "testdata/authentication_kubernetes/successful_client",
			wantErr:      false,
		},
		{
			name:         "Login should error out if the role is invalid",
			recorderMode: recorder.ModeReplaying,
			config: KubernetesConfig{
				JWTPath: "testdata/authentication_kubernetes/successful_client_token",
				Role:    "invalid-role",
			},
			playbackFile: "testdata/authentication_kubernetes/invalid_role",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)

			r, err := recorder.NewAsMode(tt.playbackFile, tt.recorderMode, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer r.Stop()
			r.SetMatcher(matchVaultResponse)

			apiConfig := api.DefaultConfig()
			apiConfig.HttpClient.Transport = r

			client, _ := api.NewClient(apiConfig)
			client.SetAddress("http://127.0.0.1:8200/")

			c := TokenKubernetesAuthentication(tt.config)
			got, err := c.Login(client)

			if tt.wantErr {
				g.Expect(err).NotTo(gomega.Succeed(), `"Login() error = %v, should have failed`, err)
			} else {
				g.Expect(err).To(gomega.Succeed(), `"Login() error = %v, should have passed"`, err)
				g.Expect(got).NotTo(gomega.BeEmpty(), `Login() should have been returned a token`)
			}
		})
	}
}
