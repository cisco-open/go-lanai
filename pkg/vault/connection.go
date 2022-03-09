package vault

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/vault/api"
)


type Client struct {
	*api.Client
	config               *ConnectionProperties
	clientAuthentication ClientAuthentication
	hooks                []Hook
}

func NewClient(p *ConnectionProperties) (*Client, error) {
	clientAuth := newClientAuthentication(p)

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

	token, err := clientAuth.Login()
	if err != nil {
		logger.Warnf("vault apiClient cannot get token %v", err)
	}
	client.SetToken(token)

	return &Client{
		Client:            client,
		config:               p,
		clientAuthentication: clientAuth,
	}, nil
}

func (c *Client) AddHooks(_ context.Context, hooks ...Hook) {
	c.hooks = append(c.hooks, hooks...)
}

func (c *Client) Logical(ctx context.Context) *Logical {
	return &Logical{
		Logical: c.Client.Logical(),
		ctx: ctx,
		client: c,
	}
}

func (c *Client) Sys(ctx context.Context) *Sys {
	return &Sys{
		Sys: c.Client.Sys(),
		ctx: ctx,
		client: c,
	}
}

func (c *Client) GetClientTokenRenewer() (*api.Renewer,  error) {
	secret, err := c.Client.Auth().Token().LookupSelf()
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
	return c.Client.NewLifetimeWatcher(&api.LifetimeWatcherInput{
		Secret: &api.Secret{
			Auth: &api.SecretAuth{
				ClientToken: c.Client.Token(),
				Renewable:   renewable,
			},
		},
		Increment: int(increment),
	})
}

func (c *Client) MonitorRenew(ctx context.Context, r *api.Renewer, renewerDescription string) {
	for {
		select {
		case err := <-r.DoneCh():
			if err != nil {
				logger.WithContext(ctx).Errorf("%s renewer failed %v", renewerDescription, err)
			}
			logger.WithContext(ctx).Infof("%s renewer stopped", renewerDescription)
			break
		case renewal := <-r.RenewCh():
			logger.WithContext(ctx).Infof("%s successfully renewed at %v", renewerDescription, renewal.RenewedAt)
		}
	}
}