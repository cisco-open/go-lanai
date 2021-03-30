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
	prefixAccessFromRefreshToken        = "AR"
	prefixRefreshTokenFromSessionId     = "RS"
	prefixAccessTokenFromSessionId      = "AS"

	/*
		Original comment form Java implementation:
		When specific token of a client is used (nfv-client), we look up the session and update
		its last requested time
		These records should have an expiry time equal to the token's expiry time
	 */
	// those relationships are not needed anymore, because session ID is stored in UserAuthentication and context details
	// which are mapped from refresh and access token
	//prefixRefreshTokenToSessionId = "R_TO_S"
	//prefixAccessTokenToSessionId = "A_TO_S"

	/*
	 * Original comment form Java implementation:
	 * We also want to store the original OAuth2 Request, because JWT token doesn't carry all information
	 * from OAuth2 request (we don't want super long JWT). We don't want to carry it in SecurityContextDetails
	 * because original OAuth2 request is only needed by authorization server
	 */
	// Those relationships are not needed anymore, because addtional details such as session ID is now carried in
	// security.KeyValueDetails
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
	switch t := key.(type) {
	case oauth2.AccessToken:
		return r.loadDetailsFromAccessToken(c, t)
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

	switch t := key.(type) {
	case oauth2.AccessToken:
		return r.saveAccessTokenToDetails(c, t, details)
	default:
		return fmt.Errorf("unsupported key type %T", key)
	}
}

func (r *RedisContextDetailsStore) RemoveContextDetails(c context.Context, key interface{}) error {
	switch t := key.(type) {
	case oauth2.AccessToken:
		return r.doRemoveDetials(c, t, "")
	default:
		return fmt.Errorf("unsupported key type %T", key)
	}
}

func (r *RedisContextDetailsStore) ContextDetailsExists(c context.Context, key interface{}) bool {
	switch t := key.(type) {
	case oauth2.AccessToken:
		return r.exists(c, keyFuncAccessTokenToDetails(uniqueTokenKey(t)) )
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
// 		- RefreshToken -> SessionId			"RS"
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
//  	- AccessToken -> SessionId 		"AS"
//		- RefreshToken <-> AccessToken 	"AR"
func (r *RedisContextDetailsStore) RegisterAccessToken(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) error {
	if e := r.saveAccessTokenFromUserClient(ctx, token, oauth); e != nil {
		return e
	}

	if e := r.saveAccessTokenToSession(ctx, token, oauth); e != nil {
		return e
	}

	if e := r.saveAccessRefreshTokenRelation(ctx, token); e != nil {
		return e
	}

	return nil
}

func (r *RedisContextDetailsStore) ReadStoredAuthorization(c context.Context, token oauth2.RefreshToken) (oauth2.Authentication, error) {
	return r.loadAuthFromRefreshToken(c, token)
}

func (r *RedisContextDetailsStore) FindSessionId(ctx context.Context, token oauth2.Token) (string, error) {
	switch t := token.(type) {
	case oauth2.AccessToken:
		return r.loadSessionId(ctx, keyFuncAccessTokenFromSession(uniqueTokenKey(t), "*"))
	case oauth2.RefreshToken:
		return r.loadSessionId(ctx, keyFuncRefreshTokenFromSession(uniqueTokenKey(t), "*"))
	default:
		return "", fmt.Errorf("unsupported key type %T", token)
	}
}

// RevokeRefreshToken remove redis records:
// 		- RefreshToken -> Authentication 	"ART"
//		- RefreshToken <- User & Client 	"RUC"
// 		- RefreshToken -> SessionId			"RS"
// 		- All Access Tokens (Each implicitly remove AccessToken <-> RefreshToken "AR")
func (r *RedisContextDetailsStore) RevokeRefreshToken(ctx context.Context, token oauth2.RefreshToken) error {
	return r.doRemoveRefreshToken(ctx, token, "")
}

// RevokeAccessToken remove redis records:
//		- AccessToken -> ContextDetails	"AAT"
//		- AccessToken <- User & Client 	"AUC"
//  	- AccessToken -> SessionId 		"AS"
//		- AccessToken <-> RefreshToken 	"AR"
func (r *RedisContextDetailsStore) RevokeAccessToken(ctx context.Context, token oauth2.AccessToken) error {
	return r.doRemoveAccessToken(ctx, token, "")
}

// RevokeAllAccessTokens remove all access tokens associated with given refresh token,
// with help of AccessToken <-> RefreshToken "AR" records
func (r *RedisContextDetailsStore) RevokeAllAccessTokens(ctx context.Context, token oauth2.RefreshToken) error {
	rtk := uniqueTokenKey(token)
	return r.doRemoveAllAccessTokens(ctx, keyFuncAccessFromRefresh("*", rtk))
}

// RevokeUserAccess remove all access/refresh tokens issued to the given user,
// with help of AccessToken <- User & Client "AUC" & RefreshToken <- User & Client "RUC" records
func (r *RedisContextDetailsStore) RevokeUserAccess(ctx context.Context, username string, revokeRefreshToken bool) error {
	if revokeRefreshToken {
		if e := r.doRemoveAllRefreshTokens(ctx, keyFuncRefreshTokenFromUserAndClient("*", username, "*")); e != nil {
			return e
		}
	}

	return r.doRemoveAllAccessTokens(ctx, keyFuncAccessTokenFromUserAndClient("*", username, "*"))
}

// RevokeClientAccess remove all access/refresh tokens issued to the given client,
// with help of AccessToken <- User & Client "AUC" & RefreshToken <- User & Client "RUC" records
func (r *RedisContextDetailsStore) RevokeClientAccess(ctx context.Context, clientId string, revokeRefreshToken bool) error {
	if revokeRefreshToken {
		if e := r.doRemoveAllRefreshTokens(ctx, keyFuncRefreshTokenFromUserAndClient("*", "*", clientId)); e != nil {
			return e
		}
	}

	return r.doRemoveAllAccessTokens(ctx, keyFuncAccessTokenFromUserAndClient("*", "*", clientId))
}

// RevokeSessionAccess remove all access/refresh tokens issued under given session,
// with help of AccessToken <- SessionId "RS" & RefreshToken <- SessionId "RS"
func (r *RedisContextDetailsStore) RevokeSessionAccess(ctx context.Context, sessionId string, revokeRefreshToken bool) error {
	if revokeRefreshToken {
		if e := r.doRemoveAllRefreshTokens(ctx, keyFuncRefreshTokenFromSession("*", sessionId)); e != nil {
			return e
		}
	}

	return r.doRemoveAllAccessTokens(ctx, keyFuncAccessTokenFromSession("*", sessionId))
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

func (r *RedisContextDetailsStore) doDelete(c context.Context, keyFunc keyFunc) (int, error) {
	k := keyFunc(r.vTag)
	cmd := r.client.Del(c, k)
	return int(cmd.Val()), cmd.Err()
}

func (r *RedisContextDetailsStore) doDeleteWithKeys(c context.Context, keys []string) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	cmd := r.client.Del(c, keys...)
	return int(cmd.Val()), cmd.Err()
}

func (r *RedisContextDetailsStore) doDeleteWithWildcard(c context.Context, keyFunc keyFunc) (int, error) {
	keys, e := r.doList(c, keyFunc)
	if e != nil {
		return 0, e
	} else if len(keys) == 0 {
		return 0, nil
	}
	cmd := r.client.Del(c, keys...)
	return int(cmd.Val()), cmd.Err()
}

func (r *RedisContextDetailsStore) doList(c context.Context, keyFunc keyFunc) ([]string, error) {
	k := keyFunc(r.vTag)
	cmd := r.client.Keys(c, k)
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Val(), nil
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
	if e := r.doSave(c, keyFuncAccessTokenToDetails(uniqueTokenKey(t)), details, t.ExpiryTime()); e != nil {
		return e
	}

	return nil
}

func (r *RedisContextDetailsStore) saveAccessTokenFromUserClient(c context.Context, t oauth2.AccessToken, oauth oauth2.Authentication) error {
	clientId := oauth.OAuth2Request().ClientId()
	username, _ := security.GetUsername(oauth.UserAuthentication())
	atk := uniqueTokenKey(t)
	rl := internal.RelationTokenUserClient{
		Username: username,
		ClientId: clientId,
		RelationToken: 	internal.RelationToken{TokenKey:  atk},
	}
	return r.doSave(c, keyFuncAccessTokenFromUserAndClient(atk, username, clientId), &rl, t.ExpiryTime())
}

func (r *RedisContextDetailsStore) saveAccessTokenToSession(c context.Context, t oauth2.AccessToken, oauth oauth2.Authentication) error {
	sid := r.findSessionId(c, oauth)
	if sid == "" {
		return nil
	}
	atk := uniqueTokenKey(t)
	rl := internal.RelationTokenSession{
		SessionId: sid,
		RelationToken: 	internal.RelationToken{TokenKey:  atk},
	}
	return r.doSave(c, keyFuncAccessTokenFromSession(atk, sid), &rl, t.ExpiryTime())
}

func (r *RedisContextDetailsStore) saveAccessRefreshTokenRelation(c context.Context, t oauth2.AccessToken) error {
	if t.RefreshToken() == nil {
		return nil
	}

	atk := uniqueTokenKey(t)
	rtk := uniqueTokenKey(t.RefreshToken())
	rl := internal.RelationAccessRefresh{
		RelationToken: 	internal.RelationToken{TokenKey:  atk},
		RefreshTokenKey: rtk,
	}
	return r.doSave(c, keyFuncAccessFromRefresh(atk, rtk), &rl, t.ExpiryTime())
}

func (r *RedisContextDetailsStore) loadDetailsFromAccessToken(c context.Context, t oauth2.AccessToken) (security.ContextDetails, error) {
	fullDetails := internal.NewFullContextDetails()
	if e := r.doLoad(c, keyFuncAccessTokenToDetails(uniqueTokenKey(t)), &fullDetails); e != nil {
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

func (r *RedisContextDetailsStore) loadSessionId(ctx context.Context, keyfunc keyFunc) (string, error) {
	keys, e := r.doList(ctx, keyfunc)
	if e != nil{
		return "", e
	} else if len(keys) == 0 {
		return "", fmt.Errorf("session ID not found for token")
	}

	rl := internal.RelationTokenSession{}
	if e := r.doLoad(ctx, keyFuncLiteral(keys[0]), &rl); e != nil {
		return "", e
	}
	return rl.SessionId, nil
}

func (r *RedisContextDetailsStore) doRemoveDetials(ctx context.Context, token oauth2.AccessToken, atk string) error {
	if token != nil {
		atk = uniqueTokenKey(token)
	}
	if _, e := r.doDelete(ctx, keyFuncAccessTokenToDetails(atk)); e != nil {
		return e
	}
	return nil
}

//		- AccessToken -> ContextDetails	"AAT"
//		- AccessToken <- User & Client 	"AUC"
//  	- AccessToken -> SessionId 		"AS"
//		- AccessToken <-> RefreshToken 	"AR"
func (r *RedisContextDetailsStore) doRemoveAccessToken(ctx context.Context, token oauth2.AccessToken, atk string) error {
	if token != nil {
		atk = uniqueTokenKey(token)
	}
	r.doRemoveDetials(ctx, token, atk)
	r.doDeleteWithWildcard(ctx, keyFuncAccessTokenFromUserAndClient(atk, "*", "*"))
	r.doDeleteWithWildcard(ctx, keyFuncAccessTokenFromSession(atk, "*"))
	r.doDeleteWithWildcard(ctx, keyFuncAccessFromRefresh(atk, "*"))
	return nil
}

func (r *RedisContextDetailsStore) doRemoveAllAccessTokens(ctx context.Context, keyfunc keyFunc) error {
	keys, e := r.doList(ctx, keyfunc)
	if e != nil {
		return e
	}

	for _, key := range keys {
		rl := internal.RelationToken{}
		if e := r.doLoad(ctx, keyFuncLiteral(key), &rl); e != nil {
			continue
		}
		r.doRemoveAccessToken(ctx, nil, rl.TokenKey)
	}

	_, e = r.doDeleteWithKeys(ctx, keys)
	return e
}

/*********************
	Refresh Token
 *********************/
func (r *RedisContextDetailsStore) saveRefreshTokenToAuth(c context.Context, t oauth2.RefreshToken, oauth oauth2.Authentication) error {
	return r.doSave(c, keyFuncRefreshTokenToAuth(uniqueTokenKey(t)), oauth, t.ExpiryTime())
}

func (r *RedisContextDetailsStore) saveRefreshTokenFromUserClient(c context.Context, t oauth2.RefreshToken, oauth oauth2.Authentication) error {
	clientId := oauth.OAuth2Request().ClientId()
	username, _ := security.GetUsername(oauth.UserAuthentication())
	rtk := uniqueTokenKey(t)
	rl := internal.RelationTokenUserClient{
		Username: username,
		ClientId: clientId,
		RelationToken: 	internal.RelationToken{TokenKey:  rtk},
	}
	return r.doSave(c, keyFuncRefreshTokenFromUserAndClient(rtk, username, clientId), &rl, t.ExpiryTime())
}

func (r *RedisContextDetailsStore) saveRefreshTokenToSession(c context.Context, t oauth2.RefreshToken, oauth oauth2.Authentication) error {
	sid := r.findSessionId(c, oauth)
	if sid == "" {
		return nil
	}
	rtk := uniqueTokenKey(t)
	rl := internal.RelationTokenSession{
		SessionId: sid,
		RelationToken: 	internal.RelationToken{TokenKey:  rtk},
	}
	return r.doSave(c, keyFuncRefreshTokenFromSession(rtk, sid), &rl, t.ExpiryTime())
}

func (r *RedisContextDetailsStore) loadAuthFromRefreshToken(c context.Context, t oauth2.RefreshToken) (oauth2.Authentication, error) {
	oauth := oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
		opt.Request = oauth2.NewOAuth2Request()
		opt.UserAuth = oauth2.NewUserAuthentication()
		opt.Details = map[string]interface{}{}
	})
	if e := r.doLoad(c, keyFuncRefreshTokenToAuth(uniqueTokenKey(t)), &oauth); e != nil {
		return nil, e
	}
	return oauth, nil
}

// 		- RefreshToken -> Authentication 	"ART"
//		- RefreshToken <- User & Client 	"RUC"
// 		- RefreshToken -> SessionId			"RS"
// 		- All Access Tokens (Each implicitly remove AccessToken <-> RefreshToken "AR")
func (r *RedisContextDetailsStore) doRemoveRefreshToken(ctx context.Context, token oauth2.RefreshToken, rtk string) error {
	if token != nil {
		rtk = uniqueTokenKey(token)
	}
	r.doDelete(ctx, keyFuncRefreshTokenToAuth(rtk))
	r.doDeleteWithWildcard(ctx, keyFuncRefreshTokenFromUserAndClient(rtk, "*", "*"))
	r.doDeleteWithWildcard(ctx, keyFuncRefreshTokenFromSession(rtk, "*"))
	r.doRemoveAllAccessTokens(ctx, keyFuncAccessFromRefresh("*", rtk))
	return nil
}

func (r *RedisContextDetailsStore) doRemoveAllRefreshTokens(ctx context.Context, keyfunc keyFunc) error {
	keys, e := r.doList(ctx, keyfunc)
	if e != nil {
		return e
	}

	for _, key := range keys {
		rl := internal.RelationToken{}
		if e := r.doLoad(ctx, keyFuncLiteral(key), &rl); e != nil {
			continue
		}
		r.doRemoveRefreshToken(ctx, nil, rl.TokenKey)
	}

	_, e = r.doDeleteWithKeys(ctx, keys)
	return e
}

/*********************
	Other Helpers
 *********************/
func (r *RedisContextDetailsStore) findSessionId(c context.Context, oauth oauth2.Authentication) (ret string) {
	// try get it from UserAuthentaction first.
	// this should works on non-proxied authentications
	var sid interface{}
	defer func() {
		if s, ok := sid.(string); ok {
			ret = s
		}
	}()

	if userAuth, ok := oauth.UserAuthentication().(oauth2.UserAuthentication); ok && userAuth.DetailsMap() != nil {
		if sid, ok = userAuth.DetailsMap()[security.DetailsKeySessionId]; ok && sid != "" {
			return
		}
	}

	// in case of proxied authentications, this value should be carried from KeyValueDetails
	if kvs, ok := oauth.Details().(security.KeyValueDetails); ok {
		if sid, ok = kvs.Value(security.DetailsKeySessionId); ok && sid != "" {
			return
		}
	}
	return
}


/*********************
	Keys
 *********************/
type keyFunc func(tag string) string

func keyFuncLiteral(key string) keyFunc {
	return func(tag string) string {
		return key
	}
}

func keyFuncAccessTokenToDetails(atk string) keyFunc {
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s", prefixAccessTokenToDetails, tag, atk)
	}
}

func keyFuncAccessTokenFromUserAndClient(atk, username, clientId string) keyFunc {
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s:%s:%s", prefixAccessTokenFromUserAndClient, tag, username, clientId, atk)
	}
}

func keyFuncAccessTokenFromSession(atk, sid string) keyFunc {
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s:%s", prefixAccessTokenFromSessionId, tag, sid, atk)
	}
}

func keyFuncAccessFromRefresh(atk, rtk string) keyFunc {
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s:%s", prefixAccessFromRefreshToken, tag, atk, rtk)
	}
}

func keyFuncRefreshTokenToAuth(rtk string) keyFunc {
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s", prefixRefreshTokenToAuthentication, tag, rtk)
	}
}

func keyFuncRefreshTokenFromUserAndClient(rtk, username, clientId string) keyFunc {
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s:%s:%s", prefixRefreshTokenFromUserAndClient, tag, username, clientId, rtk)
	}
}

func keyFuncRefreshTokenFromSession(rtk, sid string) keyFunc {
	return func(tag string) string {
		return fmt.Sprintf("%s:%s:%s:%s", prefixRefreshTokenFromSessionId, tag, sid, rtk)
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