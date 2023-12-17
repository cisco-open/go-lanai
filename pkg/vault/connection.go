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

type ClientOptions func(cfg *ClientConfig) error
type ClientConfig struct {
	*api.Config
	Properties *ConnectionProperties
	ClientAuth ClientAuthentication
	Hooks      []Hook
}

func WithProperties(p ConnectionProperties) ClientOptions {
	return func(cfg *ClientConfig) error {
		cfg.Properties = &p
		cfg.ClientAuth = newClientAuthentication(&p)
		cfg.Address = p.Address()
		if p.Scheme == "https" {
			t := api.TLSConfig{
				CACert:     p.Ssl.Cacert,
				ClientCert: p.Ssl.ClientCert,
				ClientKey:  p.Ssl.ClientKey,
				Insecure:   p.Ssl.Insecure,
			}
			err := cfg.ConfigureTLS(&t)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

type Client struct {
	*api.Client
	properties *ConnectionProperties
	clientAuth ClientAuthentication
	hooks      []Hook
}

func NewClient(opts ...ClientOptions) (*Client, error) {
	cfg := ClientConfig{
		Config:     api.DefaultConfig(),
		ClientAuth: TokenClientAuthentication(""),
	}
	for _, fn := range opts {
		if e := fn(&cfg); e != nil {
			return nil, e
		}
	}

	client, err := api.NewClient(cfg.Config)
	if err != nil {
		return nil, err
	}

	ret := &Client{
		Client:     client,
		properties: cfg.Properties,
		clientAuth: cfg.ClientAuth,
		hooks:      cfg.Hooks,
	}

	if err = ret.Authenticate(); err != nil {
		logger.Warnf("vault apiClient cannot get token %v", err)
	}
	return ret, nil
}

func (c *Client) Authenticate() error {
	token, err := c.clientAuth.Login(c.Client)
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
	props := *c.properties
	hooks := make([]Hook, len(c.hooks))
	for i := range c.hooks {
		hooks[i] = c.hooks[i]
	}

	ret := &Client{
		Client:     newClient,
		properties: &props,
		clientAuth: c.clientAuth,
		hooks:      hooks,
	}

	if e := ret.Authenticate(); e != nil {
		logger.Warnf("vault client clone cannot get token %v", e)
	}
	return ret, nil
}
