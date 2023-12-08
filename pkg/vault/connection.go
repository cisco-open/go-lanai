package vault

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/vault/api"
)

var (
	errTokenNotRenewable = errors.New("token is not renewable")
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

	ret := &Client{
		Client:               client,
		config:               p,
		clientAuthentication: clientAuth,
	}

	err = ret.Authenticate()
	if err != nil {
		logger.Warnf("vault apiClient cannot get token %v", err)
	}
	return ret, nil
}

func (c *Client) Authenticate() error {
	token, err := c.clientAuthentication.Login(c.Client)
	if err != nil {
		return err
	}
	c.Client.SetToken(token)

	return nil
}
func (c *Client) AddHooks(_ context.Context, hooks ...Hook) {
	c.hooks = append(c.hooks, hooks...)
}

func (c *Client) Logical(ctx context.Context) *Logical {
	return &Logical{
		Logical: c.Client.Logical(),
		ctx:     ctx,
		client:  c,
	}
}

func (c *Client) Sys(ctx context.Context) *Sys {
	return &Sys{
		Sys:    c.Client.Sys(),
		ctx:    ctx,
		client: c,
	}
}

func (c *Client) GetClientTokenRenewer() (*api.Renewer, error) {
	secret, err := c.Client.Auth().Token().LookupSelf()
	if err != nil {
		return nil, err
	}
	var renewable bool
	if v, ok := secret.Data["renewable"]; ok {
		renewable, _ = v.(bool)
	}
	if !renewable {
		return nil, errTokenNotRenewable
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

func (c *Client) Clone(customizers ...func(cfg *api.Config)) (*Client, error) {
	cfg := c.Client.CloneConfig()
	for _, fn := range customizers {
		fn(cfg)
	}
	newClient, e := api.NewClient(cfg)
	if e != nil {
		return nil, e
	}
	props := *c.config
	hooks := make([]Hook, len(c.hooks))
	for i := range c.hooks {
		hooks[i] = c.hooks[i]
	}

	ret := &Client{
		Client:               newClient,
		config:               &props,
		clientAuthentication: c.clientAuthentication,
		hooks:                hooks,
	}

	if e := ret.Authenticate(); e != nil {
		logger.Warnf("vault client clone cannot get token %v", e)
	}
	return ret, nil
}
