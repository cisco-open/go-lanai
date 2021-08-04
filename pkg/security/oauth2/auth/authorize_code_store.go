package auth

import (
	"context"
	"crypto/rand"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"time"
)

const (
	defaultAuthCodeLength = 32
	authCodePrefix = "AC"
)

var (
	authCodeRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	authCodeValidity = 5 * time.Minute
)

/**********************
	Abstraction
 **********************/

type AuthorizationCodeStore interface {
	GenerateAuthorizationCode(ctx context.Context, r *AuthorizeRequest, user security.Authentication) (string, error)
	ConsumeAuthorizationCode(ctx context.Context, authCode string, onetime bool) (oauth2.Authentication, error)
}

/**********************
	Redis Impl
 **********************/
type RedisAuthorizationCodeStore struct {
	redisClient redis.Client
}

func NewRedisAuthorizationCodeStore(ctx context.Context, cf redis.ClientFactory, dbIndex int) *RedisAuthorizationCodeStore {
	client, e := cf.New(ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = dbIndex
	})
	if e != nil {
		panic(e)
	}

	return &RedisAuthorizationCodeStore{
		redisClient: client,
	}
}

func (s *RedisAuthorizationCodeStore) GenerateAuthorizationCode(ctx context.Context, r *AuthorizeRequest, user security.Authentication) (string, error) {
	// TODO check code_challenge_method

	request := r.OAuth2Request()
	userAuth := ConvertToOAuthUserAuthentication(user)
	toSave := oauth2.NewAuthentication(func(conf *oauth2.AuthOption) {
		conf.Request = request
		conf.UserAuth = userAuth
	})
	code := randomString(defaultAuthCodeLength)

	if e := s.save(ctx, code, toSave); e != nil {
		return "", oauth2.NewInternalError(e)
	}
	return code, nil
}

func (s *RedisAuthorizationCodeStore) ConsumeAuthorizationCode(ctx context.Context, authCode string, onetime bool) (oauth2.Authentication, error) {
	key := s.authCodeRedisKey(authCode)
	cmd := s.redisClient.Get(ctx, key)
	if cmd.Err() != nil {
		return nil, oauth2.NewInvalidGrantError(fmt.Sprintf("code [%s] is not valid", authCode))
	}

	toLoad := oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
		opt.Request = oauth2.NewOAuth2Request()
		opt.UserAuth = oauth2.NewUserAuthentication()
		opt.Details = map[string]interface{}{}
	})

	e := json.Unmarshal([]byte(cmd.Val()), &toLoad)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(fmt.Sprintf("code [%s] is not valid", authCode), e)
	}

	if onetime {
		cmd := s.redisClient.Del(ctx, key)
		if cmd.Err() != nil {
			logger.WithContext(ctx).Warn("authorization code was not removed: " + e.Error())
		}
	}
	return toLoad, nil
}

/**********************
	Helpers
 **********************/
func (s *RedisAuthorizationCodeStore) save(ctx context.Context, code string, oauth oauth2.Authentication) error {
	key := s.authCodeRedisKey(code)
	toSave, e := json.Marshal(oauth)
	if e != nil {
		return e
	}

	cmd := s.redisClient.Set(ctx, key, toSave, authCodeValidity)
	return cmd.Err()
}

func (s *RedisAuthorizationCodeStore) authCodeRedisKey(code string) string {
	return fmt.Sprintf("%s:%s", authCodePrefix, code)
}

func randomString(length int) string {
	b := make([]byte, length)
	_, e := rand.Read(b)
	if e != nil {
		mathrand.Seed(time.Now().UnixNano())
		mathrand.Read(b) // we ignore errors
	}

	m := len(authCodeRunes)
	runes := make([]rune, length)
	for i, _ := range runes {
		runes[i] = authCodeRunes[int(b[i]) % m]
	}
	return string(runes)
}
