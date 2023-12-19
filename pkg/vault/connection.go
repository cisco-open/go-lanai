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

type Options func(cfg *ClientConfig) error
type ClientConfig struct {
	*api.Config
	Properties *ConnectionProperties
	ClientAuth ClientAuthentication
	Hooks      []Hook
}

func WithProperties(p ConnectionProperties) Options {
	return func(cfg *ClientConfig) error {
		cfg.Properties = &p
		cfg.ClientAuth = newClientAuthentication(&p)
		cfg.Address = p.Address()
		if p.Scheme == "https" {
			t := api.TLSConfig{
				CACert:     p.SSL.CaCert,
				ClientCert: p.SSL.ClientCert,
				ClientKey:  p.SSL.ClientKey,
				Insecure:   p.SSL.Insecure,
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

func New(opts ...Options) (*Client, error) {
	cfg := ClientConfig{
		Config:     api.DefaultConfig(),
		ClientAuth: TokenClientAuthentication(""),
	}
	for _, fn := range opts {
		if e := fn(&cfg); e != nil {
			return nil, e
		}
	}

	return newClient(&cfg)
}

func newClient(cfg *ClientConfig) (*Client, error) {
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

func (c *Client) Clone(opts ...Options) (*Client, error) {
	cfg := ClientConfig{
		Config:     c.Client.CloneConfig(),
		Properties: c.cloneProperties(),
		ClientAuth: c.clientAuth,
		Hooks:      make([]Hook, len(c.hooks)),
	}
	for i := range c.hooks {
		cfg.Hooks[i] = c.hooks[i]
	}
	for _, fn := range opts {
		if e := fn(&cfg); e != nil {
			return nil, e
		}
	}
	return newClient(&cfg)
}

func (c *Client) cloneProperties() *ConnectionProperties {
	if c.properties == nil {
		return nil
	}
	p := *c.properties
	return &p
}
