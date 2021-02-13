package common

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common/internal"
	"encoding/json"
	"fmt"
	"time"
)

const (
	redisDB = 13

	prefixAccessTokenToDetails = "AAT"
	prefixRefreshTokenToDetails = "ART"
	prefixAccessTokenToActiveUserAndClient = "AUC"
	prefixActiveRefreshTokenToUserAndClient = "RUC"
	prefixRefreshToAccessToken = "R_TO_A"
	prefixAccessToRefreshToken = "A_TO_R"

	/*
	  	When specific token of a client is used (nfv-client), we look up the session and update
		its last requested time
		These records should have an expiry time equal to the token's expiry time
	 */
	prefixRefreshTokenToSessionId = "R_TO_S"
	prefixAccessTokenToSessionId = "A_TO_S"

	/*
	 * We also want to store the original OAuth2 Request, because JWT token doesn't carry all information
	 * from OAuth2 request (we don't want super long JWT). We don't want to carry it in SecurityContextDetails
	 * because original OAuth2 request is only needed by authorization server
	 */
	prefixAccessTokenToRequest = "ORAT"
	prefixRefreshTokenToRequest = "ORRT"
	//prefix = ""
)

// RedisContextDetailsStore implements security.ContextDetailsStore
type RedisContextDetailsStore struct {
	vTag   string
	client redis.Client
}

func NewRedisContextDetailsStore(cf redis.ClientFactory) *RedisContextDetailsStore {
	client, e := cf.New(func(opt *redis.ClientOption) {
		opt.DbIndex = redisDB
	})
	if e != nil {
		panic(e)
	}

	return &RedisContextDetailsStore{
		vTag:   security.CompatibilityReference,
		client: client,
	}
}

func (r *RedisContextDetailsStore) ReadContextDetails(c context.Context, key interface{}) (security.ContextDetails, error) {
	switch key.(type) {
	case oauth2.AccessToken:
		return r.loadDetailsFromAccessToken(c, key.(oauth2.AccessToken))
	case oauth2.RefreshToken:
		// TODO implement this
		return nil, fmt.Errorf("loading context details using refresh token is not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported key type %T", key)
	}
}

func (r *RedisContextDetailsStore) SaveContextDetails(ctx context.Context, key interface{}, details security.ContextDetails) error {
	switch details.(type) {
	case *internal.FullContextDetails:
	case *internal.SimpleContextDetails:
	default:
		return fmt.Errorf("unsupported details type %T", details)
	}

	switch key.(type) {
	case oauth2.AccessToken:
		return r.saveAccessTokenToDetails(ctx, key.(oauth2.AccessToken), details)
	case oauth2.RefreshToken:
		return fmt.Errorf("Saving context details using refresh token is not implemented yet")
	default:
		return fmt.Errorf("unsupported key type %T", key)
	}
}

func (r *RedisContextDetailsStore) RemoveContextDetails(c context.Context, key interface{}) error {
	panic("implement me")
}

func (r *RedisContextDetailsStore) ContextDetailsExists(c context.Context, key interface{}) bool {
	switch key.(type) {
	case oauth2.AccessToken:
		return r.exists(c, keyFuncAccessTokenToDetails(key.(oauth2.AccessToken)) )
	case oauth2.RefreshToken:
		panic("Saving context details using refresh token is not implemented yet")
	default:
		return false
	}
}

/*********************
	Helpers
 *********************/
func (r *RedisContextDetailsStore) doSave(c context.Context, keyFunc keyFunc, value interface{}, expiry time.Time) error {
	v, e := json.Marshal(value)
	if e != nil {
		return e
	}

	k := keyFunc(r.vTag)
	ttl := time.Duration(redis.KeepTTL)
	now := time.Now()
	if expiry.After(now) {
		ttl = expiry.Sub(now)
	}

	status := r.client.Set(c, k, v, ttl)
	return status.Err()
}

func (r *RedisContextDetailsStore) doLoad(c context.Context, keyFunc keyFunc, value interface{}) error {
	k := keyFunc(r.vTag)
	cmd := r.client.Get(c, k)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return json.Unmarshal([]byte(cmd.Val()), value)
}

func (r *RedisContextDetailsStore) exists(c context.Context, keyFunc keyFunc) bool {
	k := keyFunc(r.vTag)
	cmd := r.client.Exists(c, k)
	return cmd.Err() == nil && cmd.Val() != 0
}

func (r *RedisContextDetailsStore) saveAccessTokenToDetails(c context.Context, t oauth2.AccessToken, details security.ContextDetails) error {
	if e := r.doSave(c, keyFuncAccessTokenToDetails(t), details, t.ExpiryTime()); e != nil {
		return e
	}

	if t.RefreshToken() != nil {
		refresh := t.RefreshToken()
		if e := r.doSave(c, keyFuncRefreshTokenToDetails(refresh), details, refresh.ExpiryTime()); e != nil {
			return e
		}
	}
	return nil
}

func (r *RedisContextDetailsStore) loadDetailsFromAccessToken(c context.Context, t oauth2.AccessToken) (security.ContextDetails, error) {
	fullDetails := internal.NewFullContextDetails()
	if e := r.doLoad(c, keyFuncAccessTokenToDetails(t), &fullDetails); e != nil {
		return nil, e
	}

	if fullDetails.User.Id == "" || fullDetails.User.Username == "" {
		// no user details, we assume it's a simple context
		return &internal.SimpleContextDetails{
			Authentication: fullDetails.Authentication,
			KV:             fullDetails.KV,
		}, nil
	}

	return fullDetails, nil
}

/*********************
	Keys
 *********************/
type keyFunc func(tag string) string

func keyFuncAccessTokenToDetails(t oauth2.AccessToken) keyFunc {
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s", prefixAccessTokenToDetails, tag, t.Value())
	}
}

func keyFuncRefreshTokenToDetails(t oauth2.RefreshToken) keyFunc {
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s", prefixRefreshTokenToDetails, tag, t.Value())
	}
}