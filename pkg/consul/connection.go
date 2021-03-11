package consul

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
	"github.com/hashicorp/consul/api"
	"errors"
	"strings"
)

var logger = log.New("Consul")

const (
	PropertyPrefix = "cloud.consul"
)

var (
	ErrNoInstances = errors.New("No matching service instances found")
)

type ConnectionProperties struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Scheme  string `json:"scheme"`
	Config  struct {
		AclToken string `json:"acl-token"`
	}
}

func (c ConnectionProperties) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type Connection struct {
	config *ConnectionProperties
	client *api.Client
}

func (c *Connection) Client() *api.Client {
	return c.client
}

func (c *Connection) Host() string {
	return c.config.Host
}

func (c *Connection) ListKeyValuePairs(ctx context.Context, path string) (results map[string]interface{}, err error) {

	queryOptions := &api.QueryOptions{}
	entries, _, err := c.client.KV().List(path, queryOptions.WithContext(context.Background()))
	if err != nil {
		return nil, err
	} else if entries == nil {
		logger.WithContext(ctx).Warnf("No appconfig retrieved from consul (%s): %s", c.Host(), path)
	} else {
		logger.WithContext(ctx).Infof("Retrieved %d configs from consul (%s): %s", len(entries), c.Host(), path)
	}

	prefix := path + "/"
	results = make(map[string]interface{})
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Key, prefix) {
			continue
		}

		propName := strings.TrimPrefix(entry.Key, prefix)

		if len(propName) > 0 {
			results[propName] = string(entry.Value)
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
		logger.WithContext(ctx).Warnf("No kv pair retrieved from consul %q: %s", c.Host(), path)
		value = nil
	} else {
		logger.WithContext(ctx).Infof("Retrieved kv pair from consul %q: %s", c.Host(), path)
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

	logger.WithContext(ctx).Infof("Stored kv pair to consul %q: %s", c.Host(), path)
	return nil
}

func NewConnection(connectionConfig *ConnectionProperties) (*Connection, error) {
	clientConfig := api.DefaultConfig()
	clientConfig.Address = connectionConfig.Address()
	clientConfig.Scheme = connectionConfig.Scheme
	clientConfig.Token = connectionConfig.Config.AclToken
	clientConfig.TLSConfig.InsecureSkipVerify = true

	if client, err := api.NewClient(clientConfig); err != nil {
		return nil, err
	} else {
		return &Connection{
			config: connectionConfig,
			client: client,
		}, nil
	}
}