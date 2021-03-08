package vault

import "context"

/*
	Vault supports many different type of backends
	The generic secret engine store arbitrary secrets.
	https://www.vaultproject.io/docs/secrets/kv/kv-v1
	This KV secrets engine does not enforce TTLs for expiration, therefore this implementation does not attempt to renew the secret's lease
 */
type GenericSecretEngine struct {
	conn *Connection
}

func (e *GenericSecretEngine) ListSecrets(ctx context.Context, path string) (results map[string]interface{}, err error) {
	results = make(map[string]interface{})

	if secrets, err := e.conn.client.Logical().Read(path); err != nil {
		return nil, err
	} else if secrets != nil {
		logger.WithContext(ctx).Infof("Retrieved %d configs from vault (%s): %s", len(secrets.Data), e.conn.config.Host, path)
		for key, val := range secrets.Data {
			results[key] = val.(string)
		}
	} else {
		logger.WithContext(ctx).Warnf("No secrets retrieved from vault (%s): %s", e.conn.config.Host, path)
	}
	return results, nil
}