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
	// HttpClient the underlying http.Client to use. Default: http.DefaultClient
	HttpClient *http.Client
	// JwkSetURL the URL of JWKSet endpoint for getting all JWKs. Default: "http://localhost:8900/auth/v2/jwks"
	// e.g. http://localhost:8900/auth/v2/jwks
	JwkSetURL string
	// JwkBaseURL the base URL of the endpoint for getting JWK by kid (without tailing slash). The actual URL would be "JwkBaseURL/<kid>".
	// (Optional) When not set (empty string), the JwkSetURL is used. Default: "http://localhost:8900/auth/v2/jwks"
	// e.g. JwkBaseURL = "http://localhost:8900/auth/v2/jwks", actual URL is "http://localhost:8900/auth/v2/jwks/<kid>"
	JwkBaseURL string
	// JwkSetRequestFunc a function that create http.Request for JWKSet endpoint. When set, override JwkSetURL.
	// (Optional) When not set, JwkSetURL is used with GET method.
	JwkSetRequestFunc func(ctx context.Context) *http.Request
	// JwkRequestFunc a function that create http.Request for "get JWK by kid". When set, override JwkBaseURL.
	// (Optional) When not set, JwkBaseURL is used with GET method. If JwkBaseURL is not set either, JWKSet endpoint is used.
	JwkRequestFunc func(ctx context.Context, kid string) *http.Request
	// DisableCache disable internal caching. If the cache is disabled, the store would invoke an external HTTP transaction
	// everytime when any of store's method is called. Default: false
	DisableCache bool
	// TTL cache setting. TTL controls how long the HTTP result is kept in cache.
	TTL time.Duration
	// RetryBackoff cache setting. It controls how long to wait between failed HTTP retries.
	RetryBackoff time.Duration
	// Retry cache setting. It controls how many times the cache would retry for failed HTTP transaction.
	Retry int
}

// NewRemoteJwkStore creates a JwkStore that load JWK with public key from an external JWKSet endpoint.
// Note: Use RemoteJwkStore with JwtDecoder ONLY.
//
//	RemoteJwkStore is not capable of decrypt private key from JWK response.
//
// See RemoteJwkStore for more details
func NewRemoteJwkStore(opts ...RemoteJwkOptions) *RemoteJwkStore {
	store := RemoteJwkStore{
		RemoteJwkConfig: RemoteJwkConfig{
			HttpClient:   http.DefaultClient,
			JwkSetURL:    "http://localhost:8900/auth/v2/jwks",
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
	if store.JwkRequestFunc == nil && len(store.JwkBaseURL) != 0 {
		store.JwkRequestFunc = remoteJwkRequestFuncWithUrl(store.JwkBaseURL)
	}
	return &store
}

// RemoteJwkStore implements JwkStore and load JWK with public key from an external JWKSet endpoint.
// Important: Use RemoteJwkStore with JwtDecoder ONLY.
//
//	RemoteJwkStore is not capable of decrypt private key from JWK response
//
// Note: LoadByName and LoadAll would treat Jwk's "name" as "kid". Because "name" is introduced for managing
//
//	key rotation, which is not applicable to JwtDecoder: JwtDecoder strictly use `kid` if present in header
//	or default "name" (in such case, should be hard coded globally known "kid")
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
	if s.JwkRequestFunc == nil {
		// JWK by kid is not available, use JWKSet endpoint
		jwks, e := s.fetchJwkSet(ctx)
		if e != nil {
			return nil, e
		}
		for _, jwk := range jwks {
			if kid == jwk.Id() {
				return jwk, nil
			}
		}
		return nil, fmt.Errorf(`failed to fetch JWK with kid [%s]: kid does not exist`, kid)
	}
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
