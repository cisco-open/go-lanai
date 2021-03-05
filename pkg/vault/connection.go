package vault

import (
	"context"
	"github.com/hashicorp/vault/api"
)


type Connection struct {
	config      *ConnectionProperties
	client      *api.Client
	tokenSource ClientAuthentication
}

func NewConnection(p *ConnectionProperties, tokenSource ClientAuthentication) (*Connection, error) {
	clientConfig := api.DefaultConfig()
	clientConfig.Address = p.Address()
	if p.Scheme == "https" {
		t := api.TLSConfig{
			CACert:     p.Ssl.Cacert,
			ClientCert: p.Ssl.ClientCert,
			ClientKey:  p.Ssl.ClientKey,
			Insecure:   p.Ssl.Insecure,
		}
		err := clientConfig.ConfigureTLS(&t)
		if err != nil {
			return nil, err
		}
	}

	client, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}

	token, err := tokenSource.Login()
	client.SetToken(token)

	return &Connection{
		config:      p,
		client:      client,
		tokenSource: tokenSource,
	}, nil
}

//TODO: secrets should be lease aware
func (c *Connection) ListSecrets(ctx context.Context, path string) (results map[string]interface{}, err error) {
	results = make(map[string]interface{})

	if secrets, err := c.client.Logical().Read(path); err != nil {
		return nil, err
	} else if secrets != nil {
		logger.WithContext(ctx).Infof("Retrieved %d configs from vault (%s): %s", len(secrets.Data), c.config.Host, path)
		for key, val := range secrets.Data {
			results[key] = val.(string)
		}
	} else {
		logger.WithContext(ctx).Warnf("No secrets retrieved from vault (%s): %s", c.config.Host, path)
	}
	return results, nil
}