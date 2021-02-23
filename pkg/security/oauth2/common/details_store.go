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
	prefixRefreshTokenToAuthentication = "ART"
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

// RedisContextDetailsStore implements security.ContextDetailsStore and auth.AuthorizationRegistry
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

/**********************************
	security.ContextDetailsStore
 **********************************/
func (r *RedisContextDetailsStore) ReadContextDetails(c context.Context, key interface{}) (security.ContextDetails, error) {
	switch key.(type) {
	case oauth2.AccessToken:
		return r.loadDetailsFromAccessToken(c, key.(oauth2.AccessToken))
	default:
		return nil, fmt.Errorf("unsupported key type %T", key)
	}
}

func (r *RedisContextDetailsStore) SaveContextDetails(c context.Context, key interface{}, details security.ContextDetails) error {
	switch details.(type) {
	case *internal.FullContextDetails:
	case *internal.SimpleContextDetails:
	default:
		return fmt.Errorf("unsupported details type %T", details)
	}

	switch key.(type) {
	case oauth2.AccessToken:
		// TODO save relationships
		return r.saveAccessTokenToDetails(c, key.(oauth2.AccessToken), details)
	default:
		return fmt.Errorf("unsupported key type %T", key)
	}
}

func (r *RedisContextDetailsStore) RemoveContextDetails(c context.Context, key interface{}) error {
	switch key.(type) {
	case oauth2.AccessToken:
		// TODO
		panic("implement me")
	default:
		return fmt.Errorf("unsupported key type %T", key)
	}
}

func (r *RedisContextDetailsStore) ContextDetailsExists(c context.Context, key interface{}) bool {
	switch key.(type) {
	case oauth2.AccessToken:
		return r.exists(c, keyFuncAccessTokenToDetails(key.(oauth2.AccessToken)) )
	default:
		return false
	}
}

/**********************************
	auth.AuthorizationRegistry
 **********************************/
func (r *RedisContextDetailsStore) ReadStoredAuthorization(c context.Context, token oauth2.RefreshToken) (oauth2.Authentication, error) {
	return r.loadAuthFromRefreshToken(c, token)
}

func (r *RedisContextDetailsStore) RegisterRefreshToken(c context.Context, token oauth2.RefreshToken, oauth oauth2.Authentication) error {
	if e := r.saveRefreshTokenToAuth(c, token, oauth); e != nil {
		return e
	}

	// TODO save relationship
	return nil
}

func (r *RedisContextDetailsStore) RefreshTokenExists(c context.Context, token oauth2.RefreshToken) bool {
	return r.exists(c, keyFuncRefreshTokenToAuth(token))
}

func (r *RedisContextDetailsStore) RevokeRefreshToken(c context.Context, token oauth2.RefreshToken) error {
	// TODO do the magic
	panic("implement me")
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

func (r *RedisContextDetailsStore) saveRefreshTokenToAuth(c context.Context, t oauth2.RefreshToken, oauth oauth2.Authentication) error {
	return r.doSave(c, keyFuncRefreshTokenToAuth(t), oauth, t.ExpiryTime())
}

func (r *RedisContextDetailsStore) loadAuthFromRefreshToken(c context.Context, t oauth2.RefreshToken) (oauth2.Authentication, error) {
	oauth := oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
		opt.Request = oauth2.NewOAuth2Request()
		opt.UserAuth = oauth2.NewUserAuthentication()
		opt.Details = map[string]interface{}{}
	})
	if e := r.doLoad(c, keyFuncRefreshTokenToAuth(t), &oauth); e != nil {
		return nil, e
	}
	return oauth, nil
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

func keyFuncRefreshTokenToAuth(t oauth2.RefreshToken) keyFunc {
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s", prefixRefreshTokenToAuthentication, tag, t.Value())
	}
}