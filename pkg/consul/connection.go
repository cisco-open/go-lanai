package consul

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"strings"
)

var logger = log.New("Consul")

const (
	PropertyPrefix = "cloud.consul"
)

var (
	ErrNoInstances = errors.New("no matching service instances found")
)

type Connection struct {
	client     *api.Client
	properties *ConnectionProperties
	clientAuth ClientAuthentication
}

func (c *Connection) Client() *api.Client {
	return c.client
}

func (c *Connection) ListKeyValuePairs(ctx context.Context, path string) (results map[string]interface{}, err error) {

	queryOptions := &api.QueryOptions{}
	entries, _, err := c.client.KV().List(path, queryOptions.WithContext(ctx))
	if err != nil {
		return nil, err
	} else if entries == nil {
		logger.WithContext(ctx).Warnf("No appconfig retrieved from consul (%s): %s", c.host(), path)
	} else {
		logger.WithContext(ctx).Infof("Retrieved %d configs from consul (%s): %s", len(entries), c.host(), path)
	}

	prefix := path + "/"
	results = make(map[string]interface{})
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Key, prefix) {
			continue
		}

		propName := strings.TrimPrefix(entry.Key, prefix)

		if len(propName) > 0 {
			strVal := string(entry.Value)
			results[propName] = utils.ParseString(strVal)
		}
	}

	if err != nil {
		return nil, err
	}
	return results, nil
}

func (c *Connection) GetKeyValue(ctx context.Context, path string) (value []byte, err error) {

	queryOptions := &api.QueryOptions{}
	data, _, err := c.client.KV().Get(path, queryOptions.WithContext(ctx))
	if err != nil {
		return nil, err
	} else if data == nil {
		logger.WithContext(ctx).Warnf("No kv pair retrieved from consul %q: %s", c.host(), path)
		value = nil
	} else {
		logger.WithContext(ctx).Infof("Retrieved kv pair from consul %q: %s", c.host(), path)
		value = data.Value
	}

	if err != nil {
		return nil, err
	}

	return
}

func (c *Connection) SetKeyValue(ctx context.Context, path string, value []byte) error {
	kvPair := &api.KVPair{
		Key:   path,
		Value: value,
	}

	writeOptions := &api.WriteOptions{}
	_, err := c.client.KV().Put(kvPair, writeOptions.WithContext(ctx))
	if err != nil {
		return err
	}

	logger.WithContext(ctx).Infof("Stored kv pair to consul %q: %s", c.host(), path)
	return nil
}

func (c *Connection) host() string {
	return fmt.Sprintf(`%s:%d`, c.properties.Host, c.properties.Port)
}

type Options func(cfg *ClientConfig) error
type ClientConfig struct {
	*api.Config
	Properties *ConnectionProperties
	ClientAuth ClientAuthentication
}

func WithProperties(p ConnectionProperties) Options {
	return func(cfg *ClientConfig) error {
		cfg.Properties = &p
		cfg.ClientAuth = newClientAuthentication(&p)
		cfg.Address = p.Address()
		cfg.Scheme = p.Scheme
		if cfg.Scheme == "https" {
			cfg.TLSConfig.CAFile = p.SSL.CaCert
			cfg.TLSConfig.CertFile = p.SSL.ClientCert
			cfg.TLSConfig.KeyFile = p.SSL.ClientKey
			cfg.TLSConfig.InsecureSkipVerify = p.SSL.Insecure
		}
		return nil
	}
}

func New(opts ...Options) (*Connection, error) {
	cfg := ClientConfig{
		Config:     api.DefaultConfig(),
		ClientAuth: TokenClientAuthentication(""),
	}
	for _, fn := range opts {
		if e := fn(&cfg); e != nil {
			return nil, e
		}
	}
	return newConn(&cfg)
}

func newConn(cfg *ClientConfig) (*Connection, error) {
	client, err := api.NewClient(cfg.Config)
	if err != nil {
		return nil, err
	}

	if cfg.ClientAuth != nil {
		token, err := cfg.ClientAuth.Login(client)
		if err != nil {
			return nil, err
		}
		cfg.Token = token
	}

	client, err = api.NewClient(cfg.Config)
	if err != nil {
		return nil, err
	}
	return &Connection{
		client:     client,
		properties: cfg.Properties,
		clientAuth: cfg.ClientAuth,
	}, nil
}
