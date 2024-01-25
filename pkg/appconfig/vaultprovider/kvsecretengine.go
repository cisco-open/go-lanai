package vaultprovider

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"errors"
	"fmt"
)

type KvSecretEngine interface {
	ContextPath(secretPath string) string
	ListSecrets(ctx context.Context, secretPath string) (results map[string]interface{}, err error)
}

func NewKvSecretEngine(version int, backend string, client *vault.Client) (KvSecretEngine, error) {
	switch version {
	case 1:
		return &KvSecretEngineV1{
			backend: backend,
			client: client,
		}, nil
	default :
		return nil, errors.New("unsupported kv secret engine version")
	}
}

type KvSecretEngineV1 struct {
	client *vault.Client
	backend  string
}

// ContextPath
//key value v1 API expects GET /secret/:path (as opposed to the v2 API which expects GET /secret/data/:path?version=:version-number)
func (engine *KvSecretEngineV1) ContextPath(secretPath string) string {
	return fmt.Sprintf("%s/%s", engine.backend, secretPath)
}

// ListSecrets implements KvSecretEngine
/*
Vault key value v1 API has the following response
we return the kv in the data field
{
  "auth": null,
  "data": {
    "foo": "bar",
    "ttl": "1h"
  },
  "lease_duration": 3600,
  "lease_id": "",
  "renewable": false
}
as opposed to the v2 API where the response is
{
  "data": {
    "data": {
      "foo": "bar"
    },
    "metadata": {
      "created_time": "2018-03-22T02:24:06.945319214Z",
      "deletion_time": "",
      "destroyed": false,
      "version": 2
    }
  }
}
*/
func (engine *KvSecretEngineV1) ListSecrets(ctx context.Context, secretPath string) (results map[string]interface{}, err error) {
	path := engine.ContextPath(secretPath)
	results = make(map[string]interface{})

	if secrets, err := engine.client.Logical(ctx).Read(path); err != nil {
		return nil, err
	} else if secrets != nil {
		for key, val := range secrets.Data {
			results[key] = utils.ParseString(val.(string))
		}
	}
	return results, nil
}