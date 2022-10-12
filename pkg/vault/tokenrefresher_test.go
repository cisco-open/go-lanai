package vault

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/httpvcr/recorder"
	"github.com/hashicorp/vault/api"
	"github.com/onsi/gomega"
	"testing"
	"time"
)

func TestTokenRefresher_Start_RefreshableToken(t *testing.T) {
	tests := []struct {
		name         string
		VaultClient  Client
		playbackFile string
		recorderMode recorder.Mode
	}{
		{
			name: "Kubernetes tokens should refresh the token when it cannot be renewed",
			VaultClient: Client{
				config: &ConnectionProperties{
					TokenSource: TokenSource{
						Source: Kubernetes,
					},
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

			tt.VaultClient.Client = createTestingAPIClient(r)
			err = tt.VaultClient.Authenticate()
			if err != nil {
				t.Fatal(err)
			}
			oldToken := tt.VaultClient.Token()
			refresher := NewTokenRefresher(&tt.VaultClient)
			refresher.Start(context.Background())
			time.Sleep(6 * time.Second)
			newToken := tt.VaultClient.Token()

			g.Expect(newToken).NotTo(gomega.Equal(oldToken),
				"Token was not refreshed, before: %v, after: %v", oldToken, newToken)

			g.Expect(refresher.renewer).NotTo(gomega.BeNil(), "Renewer nilled")
			refresher.Stop()
		})
	}
}

func TestTokenRefresher_Start_NonRefreshableToken(t *testing.T) {
	tests := []struct {
		name         string
		VaultClient  Client
		playbackFile string
		waitTimeSec  time.Duration
		recorderMode recorder.Mode
	}{
		{
			name: "Tokens with TTL should not try to refresh when renewal lease expires",
			VaultClient: Client{
				config: &ConnectionProperties{
					TokenSource: TokenSource{
						Source: Token,
					},
				},
				clientAuthentication: TokenClientAuthentication("token_10s_ttl"), // Token with TTL of 10 sec
			},
			playbackFile: "testdata/tokenrefresher/TestNotRefreshableToken",
			waitTimeSec:  10 * time.Second,
			recorderMode: recorder.ModeReplaying,
		},
		{
			name: "Static tokens (without TTL) should not try to refresh or renew",
			VaultClient: Client{
				config: &ConnectionProperties{
					TokenSource: TokenSource{
						Source: Token,
					},
				},
				clientAuthentication: TokenClientAuthentication("token_no_ttl"), // Token with no ttl - cannot be renewed
			},
			waitTimeSec:  3 * time.Second,
			playbackFile: "testdata/tokenrefresher/TestNonRenewableToken",
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

			tt.VaultClient.Client = createTestingAPIClient(r)
			err = tt.VaultClient.Authenticate()
			if err != nil {
				t.Fatal(err)
			}
			oldToken := tt.VaultClient.Token()
			refresher := NewTokenRefresher(&tt.VaultClient)
			refresher.Start(context.Background())
			time.Sleep(tt.waitTimeSec)
			newToken := tt.VaultClient.Token()

			g.Expect(newToken).To(gomega.Equal(oldToken),
				"Non-refreshable Token was refreshed, before: %v, after: %v", oldToken, newToken)
			g.Expect(refresher.renewer).To(gomega.BeNil(), "Renewer not nilled")
			refresher.Stop()
		})
	}
}

func TestTokenRefresher_Start_Stop_Restart(t *testing.T) {
	type expectedStruct struct {
		beNilAfterStop    bool
		beNilAfterRestart bool
	}
	tests := []struct {
		name         string
		VaultClient  Client
		playbackFile string
		recorderMode recorder.Mode
		expected     expectedStruct
	}{
		{
			name: "Refresher should resume token renewal if stopped & restarted",
			VaultClient: Client{
				config: &ConnectionProperties{
					TokenSource: TokenSource{
						Source: Kubernetes,
					},
				},
				clientAuthentication: TokenKubernetesAuthentication(KubernetesConfig{
					JWTPath: "testdata/tokenrefresher/auth_token",
					Role:    "devweb-app",
				}),
			},
			playbackFile: "testdata/tokenrefresher/TestStartStopAndRestart",
			recorderMode: recorder.ModeReplaying,
			expected: expectedStruct{
				beNilAfterStop:    true,
				beNilAfterRestart: false,
			},
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

			tt.VaultClient.Client = createTestingAPIClient(r)
			err = tt.VaultClient.Authenticate()
			if err != nil {
				t.Fatal(err)
			}
			refresher := NewTokenRefresher(&tt.VaultClient)
			refresher.Start(context.Background())
			time.Sleep(6 * time.Second)

			refresher.Stop()
			time.Sleep(1 * time.Second)
			renewerNilAfterStop := refresher.renewer == nil
			g.Expect(renewerNilAfterStop).To(gomega.Equal(tt.expected.beNilAfterStop),
				"Expected renewer nill after stop to be %v, got %v",
				tt.expected.beNilAfterStop,
				renewerNilAfterStop)

			refresher.Start(context.Background())
			time.Sleep(6 * time.Second)
			renewerNilAfterRestart := refresher.renewer == nil
			g.Expect(renewerNilAfterRestart).To(gomega.Equal(tt.expected.beNilAfterRestart),
				"Expected renewer nill after restart to be %v, got %v",
				tt.expected.beNilAfterRestart,
				renewerNilAfterRestart)

			refresher.Stop()
		})
	}
}

func createTestingAPIClient(r *recorder.Recorder) *api.Client {
	apiConfig := api.DefaultConfig()
	apiConfig.HttpClient.Transport = r

	apiClient, _ := api.NewClient(apiConfig)
	apiClient.SetAddress("http://127.0.0.1:8200/")
	return apiClient
}
