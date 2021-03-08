package vault

import (
	"encoding/json"
	"github.com/hashicorp/vault/api"
)


type Client struct {
	*api.Client
	config               *ConnectionProperties
	clientAuthentication ClientAuthentication
}

func NewClient(p *ConnectionProperties, clientAuthentication ClientAuthentication) (*Client, error) {
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

	token, err := clientAuthentication.Login()
	client.SetToken(token)

	return &Client{
		Client: client,
		config:               p,
		clientAuthentication: clientAuthentication,
	}, nil
}

func (c *Client) GetClientTokenRenewer() (*api.Renewer,  error) {
	secret, err := c.Auth().Token().LookupSelf()
	if err != nil {
		return nil, err
	}
	var renewable bool
	if v, ok := secret.Data["renewable"]; ok {
		renewable, _ = v.(bool)
	}
	var increment int64
	if v, ok := secret.Data["ttl"]; ok {
		if n, ok := v.(json.Number); ok {
			increment, _ = n.Int64()
		}
	}
	r, err := c.NewRenewer(&api.RenewerInput{
		Secret: &api.Secret{
			Auth: &api.SecretAuth{
				ClientToken: c.Token(),
				Renewable:   renewable,
			},
		},
		Increment: int(increment),
	})
	return r, nil
}

func (c *Client) monitorRenew(r *api.Renewer, renewerDescription string) {
	for {
		select {
		case err := <-r.DoneCh():
			if err != nil {
				logger.Errorf("%s renewer failed %v", renewerDescription, err)
			}
			logger.Infof("%s renewer stopped", renewerDescription)
			break
		case renewal := <-r.RenewCh():
			logger.Infof("%s successfully renewed at %v", renewerDescription, renewal.RenewedAt)
		}
	}
}

func (c *Client) GenericSecretEngine() *GenericSecretEngine {
	return &GenericSecretEngine{
		client: c,
	}
}