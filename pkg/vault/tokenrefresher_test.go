package vault

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/httpvcr/recorder"
	"github.com/hashicorp/vault/api"
	"github.com/onsi/gomega"
	"testing"
	"time"
)

func TestTokenRefresher_startRefresher(t *testing.T) {
	tests := []struct {
		name         string
		VaultClient  Client
		playbackFile string
		recorderMode recorder.Mode
	}{
		{
			name: "Should refresh the token when it cannot be renewed",
			VaultClient: Client{
				config: &ConnectionProperties{
					Authentication: Kubernetes,
				},
				clientAuthentication: TokenKubernetesAuthentication(KubernetesConfig{
					JWTPath: "testdata/tokenrefresher/auth_token",
					Role:    "devweb-app",
				}),
			},
			playbackFile: "testdata/tokenrefresher/TestRefreshToken",
			recorderMode: recorder.ModeReplaying,
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

			apiClient, _ := api.NewClient(apiConfig)
			apiClient.SetAddress("http://127.0.0.1:8200/")
			tt.VaultClient.Client = apiClient
			err = tt.VaultClient.Authenticate()
			if err != nil {
				t.Fatal(err)
			}
			oldToken := tt.VaultClient.Token()
			refresher := NewTokenRefresher(&tt.VaultClient)
			go refresher.Start(context.Background())
			time.Sleep(6 * time.Second)
			newToken := tt.VaultClient.Token()

			g.Expect(newToken).NotTo(gomega.Equal(oldToken),
				"Token was not refreshed, before: %v, after: %v", oldToken, newToken)
			refresher.Stop()
		})
	}
}
