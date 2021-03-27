package common

import (
	"context"
	"crypto/sha256"
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

	prefixAccessTokenToDetails          = "AAT"
	prefixRefreshTokenToAuthentication  = "ART"
	prefixAccessTokenFromUserAndClient  = "AUC"
	prefixRefreshTokenFromUserAndClient = "RUC"
	prefixRefreshToAccessToken          = "R_TO_A"
	prefixAccessToRefreshToken          = "A_TO_R"

	/*
	  	When specific token of a client is used (nfv-client), we look up the session and update
		its last requested time
		These records should have an expiry time equal to the token's expiry time
	 */
	prefixRefreshTokenToSessionId = "R_TO_S"
	prefixAccessTokenToSessionId = "A_TO_S"

	/*
	 * Original comment form Java implementation:
	 * We also want to store the original OAuth2 Request, because JWT token doesn't carry all information
	 * from OAuth2 request (we don't want super long JWT). We don't want to carry it in SecurityContextDetails
	 * because original OAuth2 request is only needed by authorization server
	 */
	// Those relationships are not needed anymore, because
	//prefixAccessTokenToRequest = "ORAT"
	//prefixRefreshTokenToRequest = "ORRT"

	//prefix = ""
)

// RedisContextDetailsStore implements security.ContextDetailsStore and auth.AuthorizationRegistry
type RedisContextDetailsStore struct {
	vTag   string
	client redis.Client
}

func NewRedisContextDetailsStore(ctx context.Context, cf redis.ClientFactory) *RedisContextDetailsStore {
	client, e := cf.New(ctx, func(opt *redis.ClientOption) {
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
		// TODO implement me
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
// RegisterRefreshToken save relationships :
// 		- RefreshToken -> Authentication 	"ART"
//		- RefreshToken <- User & Client 	"RUC"
// 		- RefreshToken -> SessionId			"R_TO_S"
func (r *RedisContextDetailsStore) RegisterRefreshToken(c context.Context, token oauth2.RefreshToken, oauth oauth2.Authentication) error {
	if e := r.saveRefreshTokenToAuth(c, token, oauth); e != nil {
		return e
	}

	if e := r.saveRefreshTokenFromUserClient(c, token, oauth); e != nil {
		return e
	}

	if e := r.saveRefreshTokenToSession(c, token, oauth); e != nil {
		return e
	}
	return nil
}

// RegisterAccessToken save relationships :
//		- AccessToken <- User & Client 	"AUC"
//  	- AccessToken -> SessionId 		"A_TO_S"
//		- RefreshToken -> AccessToken 	"R_TO_A"
//		- AccessToken -> RefreshToken 	"A_TO_R"
func (r *RedisContextDetailsStore) RegisterAccessToken(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) error {
	if e := r.saveAccessTokenFromUserClient(ctx, token, oauth); e != nil {
		return e
	}

	if e := r.saveAccessTokenToSession(ctx, token, oauth); e != nil {
		return e
	}

	if e := r.saveAccessToRefreshToken(ctx, token); e != nil {
		return e
	}

	if e := r.saveRefreshToAccessToken(ctx, token); e != nil {
		return e
	}

	return nil
}

func (r *RedisContextDetailsStore) ReadStoredAuthorization(c context.Context, token oauth2.RefreshToken) (oauth2.Authentication, error) {
	return r.loadAuthFromRefreshToken(c, token)
}

func (r *RedisContextDetailsStore) RefreshTokenExists(c context.Context, token oauth2.RefreshToken) bool {
	return r.exists(c, keyFuncRefreshTokenToAuth(token))
}

func (r *RedisContextDetailsStore) RevokeRefreshToken(c context.Context, token oauth2.RefreshToken) error {
	// TODO do the magic
	panic("implement me")
}

/*********************
	Common Helpers
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

/*********************
	Access Token
 *********************/
func (r *RedisContextDetailsStore) saveAccessTokenToDetails(c context.Context, t oauth2.AccessToken, details security.ContextDetails) error {
	if e := r.doSave(c, keyFuncAccessTokenToDetails(t), details, t.ExpiryTime()); e != nil {
		return e
	}

	return nil
}

func (r *RedisContextDetailsStore) saveAccessTokenFromUserClient(c context.Context, t oauth2.AccessToken, oauth oauth2.Authentication) error {
	clientId := oauth.OAuth2Request().ClientId()
	username, _ := security.GetUsername(oauth.UserAuthentication())
	return r.doSave(c, keyFuncAccessTokenFromUserAndClient(t, username, clientId), t.Value(), t.ExpiryTime())
}

func (r *RedisContextDetailsStore) saveAccessTokenToSession(c context.Context, t oauth2.AccessToken, oauth oauth2.Authentication) error {
	sid := r.findSessionId(c, oauth)
	if sid == nil || sid == "" {
		return nil
	}
	return r.doSave(c, keyFuncAccessTokenToSession(t), sid, t.ExpiryTime())
}

func (r *RedisContextDetailsStore) saveAccessToRefreshToken(c context.Context, t oauth2.AccessToken) error {
	if t.RefreshToken() == nil {
		return nil
	}
	return r.doSave(c, keyFuncAccessToRefresh(t), t.RefreshToken().Value(), t.ExpiryTime())
}

func (r *RedisContextDetailsStore) saveRefreshToAccessToken(c context.Context, t oauth2.AccessToken) error {
	if t.RefreshToken() == nil {
		return nil
	}
	return r.doSave(c, keyFuncRefreshToAccess(t.RefreshToken()), t.Value(), t.ExpiryTime())
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
	Refresh Token
 *********************/
func (r *RedisContextDetailsStore) saveRefreshTokenToAuth(c context.Context, t oauth2.RefreshToken, oauth oauth2.Authentication) error {
	return r.doSave(c, keyFuncRefreshTokenToAuth(t), oauth, t.ExpiryTime())
}

func (r *RedisContextDetailsStore) saveRefreshTokenFromUserClient(c context.Context, t oauth2.RefreshToken, oauth oauth2.Authentication) error {
	clientId := oauth.OAuth2Request().ClientId()
	username, _ := security.GetUsername(oauth.UserAuthentication())
	return r.doSave(c, keyFuncRefreshTokenFromUserAndClient(t, username, clientId), t.Value(), t.ExpiryTime())
}

func (r *RedisContextDetailsStore) saveRefreshTokenToSession(c context.Context, t oauth2.RefreshToken, oauth oauth2.Authentication) error {
	sid := r.findSessionId(c, oauth)
	if sid == nil || sid == "" {
		return nil
	}
	return r.doSave(c, keyFuncRefreshTokenToSession(t), sid, t.ExpiryTime())
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
	Other Helpers
 *********************/
func (r *RedisContextDetailsStore) findSessionId(c context.Context, oauth oauth2.Authentication) interface{} {
	// try get it from UserAuthentaction first.
	// this should works on non-proxied authentications
	if userAuth, ok := oauth.UserAuthentication().(oauth2.UserAuthentication); ok && userAuth.DetailsMap() != nil {
		if sid, ok := userAuth.DetailsMap()[security.DetailsKeySessionId]; ok && sid != "" {
			return sid
		}
	}

	// in case of proxied authentications, this value should be carried from KeyValueDetails
	if kvs, ok := oauth.Details().(security.KeyValueDetails); ok {
		if sid, ok := kvs.Value(security.DetailsKeySessionId); ok && sid != "" {
			return sid
		}
	}
	return nil
}


/*********************
	Keys
 *********************/
type keyFunc func(tag string) string

func keyFuncAccessTokenToDetails(t oauth2.AccessToken) keyFunc {
	tk := uniqueTokenKey(t)
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s", prefixAccessTokenToDetails, tag, tk)
	}
}

func keyFuncAccessTokenFromUserAndClient(t oauth2.AccessToken, username, clientId string) keyFunc {
	tk := uniqueTokenKey(t)
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s:%s:%s", prefixAccessTokenFromUserAndClient, tag, username, clientId, tk)
	}
}

func keyFuncAccessTokenToSession(t oauth2.AccessToken) keyFunc {
	tk := uniqueTokenKey(t)
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s", prefixAccessTokenToSessionId, tag, tk)
	}
}

func keyFuncAccessToRefresh(t oauth2.AccessToken) keyFunc {
	tk := uniqueTokenKey(t)
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s", prefixAccessToRefreshToken, tag, tk)
	}
}

func keyFuncRefreshToAccess(t oauth2.RefreshToken) keyFunc {
	tk := uniqueTokenKey(t)
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s", prefixRefreshToAccessToken, tag, tk)
	}
}

func keyFuncRefreshTokenToAuth(t oauth2.RefreshToken) keyFunc {
	tk := uniqueTokenKey(t)
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s", prefixRefreshTokenToAuthentication, tag, tk)
	}
}

func keyFuncRefreshTokenFromUserAndClient(t oauth2.RefreshToken, username, clientId string) keyFunc {
	tk := uniqueTokenKey(t)
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s:%s:%s", prefixRefreshTokenFromUserAndClient, tag, username, clientId, tk)
	}
}

func keyFuncRefreshTokenToSession(t oauth2.RefreshToken) keyFunc {
	tk := uniqueTokenKey(t)
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s", prefixRefreshTokenToSessionId, tag, tk)
	}
}

func uniqueTokenKey(token oauth2.Token) string {
	// use JTI if possible
	if t, ok := token.(oauth2.ClaimsContainer); ok && t.Claims() != nil {
		if jti, ok := t.Claims().Get(oauth2.ClaimJwtId).(string); ok && jti != "" {
			return jti
		}
	}

	// use a hash of value
	hash := sha256.Sum224([]byte(token.Value()))
	return fmt.Sprintf("%x", hash)
}