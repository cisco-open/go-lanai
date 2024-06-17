package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/utils/cacheutils"
	"net/http"
	"time"
)

var ckJwkSet = cacheutils.StringKey(`github.com/cisco-open/go-lanai/JWKSet`)

type jwkSet struct {
	Keys []*GenericJwk `json:"keys"`
}

type RemoteJwkOptions func(cfg *RemoteJwkConfig)
type RemoteJwkConfig struct {
	HttpClient        *http.Client
	JwkSetURL         string
	JwkBaseURL        string
	JwkSetRequestFunc func(ctx context.Context) *http.Request
	JwkRequestFunc    func(ctx context.Context, kid string) *http.Request
	DisableCache      bool
	TTL               time.Duration
	RetryBackoff      time.Duration
	Retry             int
}

// NewRemoteJwkStore creates a JwkStore that load JWK with public key from an external JWKSet endpoint.
// Note: Use RemoteJwkStore with JwtDecoder ONLY.
//       RemoteJwkStore is not capable of decrypt private key from JWK response.
// See RemoteJwkStore for more details
func NewRemoteJwkStore(opts ...RemoteJwkOptions) *RemoteJwkStore {
	store := RemoteJwkStore{
		RemoteJwkConfig: RemoteJwkConfig{
			HttpClient:   http.DefaultClient,
			JwkSetURL:    "http://localhost:8900/auth/v2/jwks",
			JwkBaseURL:   "http://localhost:8900/auth/v2/jwks",
			TTL:          60 * time.Minute,
			RetryBackoff: 2 * time.Second,
			Retry:        2,
		},
	}
	for _, fn := range opts {
		fn(&store.RemoteJwkConfig)
	}
	if !store.DisableCache {
		store.cache = cacheutils.NewMemCache(func(opt *cacheutils.CacheOption) {
			opt.Heartbeat = store.TTL
			opt.LoadRetry = store.Retry
		})
	}
	if store.JwkSetRequestFunc == nil {
		store.JwkSetRequestFunc = remoteJwkSetRequestFuncWithUrl(store.JwkSetURL)
	}
	if store.JwkRequestFunc == nil {
		store.JwkRequestFunc = remoteJwkRequestFuncWithUrl(store.JwkBaseURL)
	}
	return &store
}

// RemoteJwkStore implements JwkStore and load JWK with public key from an external JWKSet endpoint.
// Important: Use RemoteJwkStore with JwtDecoder ONLY.
//            RemoteJwkStore is not capable of decrypt private key from JWK response
// Note: LoadByName and LoadAll would treat Jwk's "name" as "kid". Because "name" is introduced for managing
//       key rotation, which is not applicable JwtDecoder: JwtDecoder strictly use `kid` if present in header
//       or default "name" (in such case, should be hard coded globally known "kid")
type RemoteJwkStore struct {
	RemoteJwkConfig
	cache cacheutils.MemCache
}

func (s *RemoteJwkStore) LoadByKid(ctx context.Context, kid string) (Jwk, error) {
	if s.DisableCache || s.cache == nil {
		return s.fetchJwkByKid(ctx, kid)
	}
	i, e := s.cache.GetOrLoad(ctx, cacheutils.StringKey(kid), s.loadJwkByKid, nil)
	if e != nil {
		return nil, e
	}
	return i.(Jwk), nil
}

func (s *RemoteJwkStore) LoadByName(ctx context.Context, name string) (Jwk, error) {
	// Note: remote JWK endpoint doesn't give name, we treat name as KID
	if s.DisableCache || s.cache == nil {
		return s.fetchJwkByKid(ctx, name)
	}
	i, e := s.cache.GetOrLoad(ctx, cacheutils.StringKey(name), s.loadJwkByKid, nil)
	if e != nil {
		return nil, e
	}
	return i.(Jwk), nil
}

func (s *RemoteJwkStore) LoadAll(ctx context.Context, names ...string) ([]Jwk, error) {
	var loaded interface{}
	var err error
	if s.DisableCache || s.cache == nil {
		loaded, err = s.fetchJwkSet(ctx)
	} else {
		loaded, err = s.cache.GetOrLoad(ctx, ckJwkSet, s.loadJwkSet, nil)

	}
	if err != nil {
		return nil, err
	}
	return s.filterJwkSet(loaded.([]Jwk), names...), nil
}

func (s *RemoteJwkStore) loadJwkByKid(ctx context.Context, k cacheutils.Key) (v interface{}, exp time.Time, err error) {
	key := k.(cacheutils.StringKey)
	jwk, e := s.fetchJwkByKid(ctx, string(key))
	if e != nil {
		return nil, time.Now().Add(s.RetryBackoff), e
	}
	return jwk, time.Now().Add(s.TTL), nil
}

func (s *RemoteJwkStore) loadJwkSet(ctx context.Context, _ cacheutils.Key) (v interface{}, exp time.Time, err error) {
	jwks, e := s.fetchJwkSet(ctx)
	if e != nil {
		return nil, time.Now().Add(s.RetryBackoff), e
	}
	return jwks, time.Now().Add(s.TTL), nil
}

func (s *RemoteJwkStore) fetchJwkByKid(ctx context.Context, kid string) (Jwk, error) {
	req := s.JwkRequestFunc(ctx, kid)
	if req == nil {
		return nil, fmt.Errorf(`unable to resolve HTTP request for JWK with kid [%s]`, kid)
	}
	resp, e := s.doFetch(req)
	if e != nil {
		return nil, fmt.Errorf(`failed to fetch JWK with kid [%s]: %v`, kid, e)
	}
	defer func() { _ = resp.Body.Close() }()
	var jwk GenericJwk
	if e := json.NewDecoder(resp.Body).Decode(&jwk); e != nil {
		return nil, fmt.Errorf(`unable to parse JWK from JSON: %v`, e)
	}
	return &jwk, nil
}

func (s *RemoteJwkStore) fetchJwkSet(ctx context.Context) ([]Jwk, error) {
	req := s.JwkSetRequestFunc(ctx)
	if req == nil {
		return nil, fmt.Errorf(`unable to resolve HTTP request for JWK Set`)
	}
	resp, e := s.doFetch(req)
	if e != nil {
		return nil, fmt.Errorf(`failed to fetch JWK Set: %v`, e)
	}
	defer func() { _ = resp.Body.Close() }()
	var jwkSet jwkSet
	if e := json.NewDecoder(resp.Body).Decode(&jwkSet); e != nil {
		return nil, fmt.Errorf(`unable to parse JWK Set from JSON: %v`, e)
	}
	jwks := make([]Jwk, len(jwkSet.Keys))
	for i := range jwkSet.Keys {
		jwks[i] = jwkSet.Keys[i]
	}
	return jwks, nil
}

func (s *RemoteJwkStore) doFetch(req *http.Request) (*http.Response, error) {
	switch resp, e := s.HttpClient.Do(req); {
	case e != nil:
		return nil, e
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		return nil, fmt.Errorf(`failed with status code [%d: %s]`, resp.StatusCode, resp.Status)
	default:
		return resp, nil
	}
}

func (s *RemoteJwkStore) filterJwkSet(jwks []Jwk, kids ...string) []Jwk {
	if len(kids) == 0 {
		return jwks
	}
	filtered := make([]Jwk, 0, len(jwks))
	for i := range jwks {
		for _, kid := range kids {
			if jwks[i].Id() == kid {
				filtered = append(filtered, jwks[i])
				break
			}
		}
	}
	return filtered
}

func remoteJwkSetRequestFuncWithUrl(base string) func(ctx context.Context) *http.Request {
	return func(ctx context.Context) *http.Request {
		req, e := http.NewRequestWithContext(ctx, http.MethodGet, base, nil)
		if e != nil {
			return nil
		}
		return req
	}
}

func remoteJwkRequestFuncWithUrl(base string) func(ctx context.Context, kid string) *http.Request {
	return func(ctx context.Context, kid string) *http.Request {
		req, e := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(`%s/%s`, base, kid), nil)
		if e != nil {
			return nil
		}
		return req
	}
}
