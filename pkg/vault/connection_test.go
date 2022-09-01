package vault

import (
	"github.com/hashicorp/vault/api"
	"testing"
)

func TestClient_Authenticate(t *testing.T) {
	type fields struct {
		Client               *api.Client
		config               *ConnectionProperties
		clientAuthentication ClientAuthentication
		hooks                []Hook
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		expected string
	}{
		{
			name: "Authenticate should login and set the token in the client",
			fields: fields{
				clientAuthentication: TokenClientAuthentication("token_value"),
			},
			wantErr:  false,
			expected: "token_value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := api.NewClient(api.DefaultConfig())
			c := &Client{
				clientAuthentication: tt.fields.clientAuthentication,
				Client:               client,
			}
			if err := c.Authenticate(); (err != nil) != tt.wantErr {
				t.Errorf("Authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := c.Client.Token()
			if got != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, got)
			}
		})
	}
}
