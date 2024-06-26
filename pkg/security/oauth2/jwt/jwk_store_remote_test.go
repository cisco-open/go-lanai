package jwt_test

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/jwt"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/ittest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

const (
	TestRemoteKid = `dev-0`
)

/*************************
	Test Cases
 *************************/

type RemoteTestDI struct {
	fx.In
	ittest.RecorderDI
}

func TestRemoteJwkStore(t *testing.T) {
	var di RemoteTestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		ittest.WithHttpPlayback(t,
			//ittest.HttpRecordingMode(), // uncomment for record mode with httpPlayback
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestRemoteJwkStore(&di, true, true), "JwkByKidEpWithCache"),
		test.GomegaSubTest(SubTestRemoteJwkStore(&di, true, false), "JwkByKidEpWithoutCache"),
		test.GomegaSubTest(SubTestRemoteJwkStore(&di, false, true), "NoJwkByKidWithCache"),
		test.GomegaSubTest(SubTestRemoteJwkStore(&di, false, false), "NoJwkByKidWithoutCache"),

		test.GomegaSubTest(SubTestRemoteJwkStoreWithBadServer(&di, true, true), "JwkByKidEpBadServerWithCache"),
		test.GomegaSubTest(SubTestRemoteJwkStoreWithBadServer(&di, true, false), "JwkByKidEpBadServerWithoutCache"),
		test.GomegaSubTest(SubTestRemoteJwkStoreWithBadServer(&di, false, true), "NoJwkByKidBadServerWithCache"),
		test.GomegaSubTest(SubTestRemoteJwkStoreWithBadServer(&di, false, false), "NoJwkByKidBadServerWithoutCache"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestRemoteJwkStore(di *RemoteTestDI, byKid bool, cache bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		store := jwt.NewRemoteJwkStore(func(cfg *jwt.RemoteJwkConfig) {
			cfg.HttpClient = di.Recorder.GetDefaultClient()
			cfg.DisableCache = !cache
			cfg.JwkSetURL = "http://localhost:8900/auth/v2/jwks"
			if byKid {
				cfg.JwkBaseURL = "http://localhost:8900/auth/v2/jwks"
			}
		})

		test.RunTest(ctx, t,
			test.GomegaSubTest(SubTestRemoteJwk(store, "kid", false), "LoadByKid"),
			test.GomegaSubTest(SubTestRemoteNonExistJwk(store, "kid"), "LoadByNonExistKid"),
			test.GomegaSubTest(SubTestRemoteEmptyJwk(store, "kid"), "LoadByEmptyKid"),
			test.GomegaSubTest(SubTestRemoteJwk(store, "name", false), "LoadByKid"),
			test.GomegaSubTest(SubTestRemoteNonExistJwk(store, "name"), "LoadByNonExistName"),
			test.GomegaSubTest(SubTestRemoteEmptyJwk(store, "name"), "LoadByEmptyName"),
			test.GomegaSubTest(SubTestRemoteJwkSet(store, false, false), "LoadAll"),
			test.GomegaSubTest(SubTestRemoteJwkSet(store, false, false, TestRemoteKid), "LoadAllWithName"),
			test.GomegaSubTest(SubTestRemoteJwkSet(store, false, true, "non-exist"), "LoadAllWithNonExistName"),
			test.GomegaSubTest(SubTestRemoteJwkSet(store, false, true, ""), "LoadAllWithEmptyName"),
		)
	}
}

func SubTestRemoteJwkStoreWithBadServer(di *RemoteTestDI, byKid bool, cache bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		store := jwt.NewRemoteJwkStore(func(cfg *jwt.RemoteJwkConfig) {
			cfg.HttpClient = di.Recorder.GetDefaultClient()
			cfg.DisableCache = !cache
			cfg.JwkSetRequestFunc = BadRemoteJwkSetRequestFunc()
			if byKid {
				cfg.JwkRequestFunc = BadRemoteJwkRequestFunc()
			}
		})

		test.RunTest(ctx, t,
			test.GomegaSubTest(SubTestRemoteJwk(store, "kid", true), "LoadByKid"),
			test.GomegaSubTest(SubTestRemoteNonExistJwk(store, "kid"), "LoadByNonExistKid"),
			test.GomegaSubTest(SubTestRemoteEmptyJwk(store, "kid"), "LoadByEmptyKid"),
			test.GomegaSubTest(SubTestRemoteJwk(store, "name", true), "LoadByKid"),
			test.GomegaSubTest(SubTestRemoteNonExistJwk(store, "name"), "LoadByNonExistName"),
			test.GomegaSubTest(SubTestRemoteEmptyJwk(store, "name"), "LoadByEmptyName"),
			test.GomegaSubTest(SubTestRemoteJwkSet(store, true, true), "LoadAll"),
		)
	}
}

func SubTestRemoteJwk(store *jwt.RemoteJwkStore, by string, shouldFail bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const attempts = 2
		for i := 0; i < attempts; i++ {
			jwk, e := LoadRemoteJwk(ctx, store, by, TestRemoteKid)
			if shouldFail {
				g.Expect(e).To(HaveOccurred(), "load by %s [%d] should fail ", by, i)
				continue
			}
			g.Expect(e).To(Succeed(), "load by %s [%d] should not fail ", by, i)
			g.Expect(jwk).ToNot(BeZero(), "load by %s [%d] should return non-zero result", by, i)
			g.Expect(jwk).To(HaveField("Public()", Not(BeNil())), "load by %s [%d] should return non-zero public key", by, i)
		}
	}
}

func SubTestRemoteNonExistJwk(store *jwt.RemoteJwkStore, by string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const attempts = 2
		for i := 0; i < attempts; i++ {
			_, e := LoadRemoteJwk(ctx, store, by, "non-exist")
			g.Expect(e).To(HaveOccurred(), "load by %s [%d] should fail ", by, i)
		}
	}
}

func SubTestRemoteEmptyJwk(store *jwt.RemoteJwkStore, by string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const attempts = 2
		for i := 0; i < attempts; i++ {
			_, e := LoadRemoteJwk(ctx, store, by, "")
			g.Expect(e).To(HaveOccurred(), "load by %s [%d] should fail ", by, i)
		}
	}
}

func SubTestRemoteJwkSet(store *jwt.RemoteJwkStore, shouldFail, expectEmpty bool, filterBy...string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const attempts = 2
		for i := 0; i < attempts; i++ {
			jwks, e := store.LoadAll(ctx, filterBy...)
			if shouldFail {
				g.Expect(e).To(HaveOccurred(), "LoadAll [%d] should fail ", i)
				continue
			}

			g.Expect(e).To(Succeed(), "LoadAll [%d] should not fail ", i)
			if expectEmpty {
				g.Expect(jwks).To(BeEmpty(), "LoadAll [%d] should return empty result", i)
			} else {
				g.Expect(jwks).ToNot(BeEmpty(), "LoadAll [%d] should return non-empty result", i)
				g.Expect(jwks).To(HaveEach(HaveField("Public()", Not(BeNil()))),
					"LoadAll [%d] should return JWK set with valid public keys", i)
			}
		}
	}
}

/*************************
	Helpers
 *************************/

func LoadRemoteJwk(ctx context.Context, store *jwt.RemoteJwkStore, by, val string) (jwt.Jwk, error) {
	switch by {
	case "kid":
		return store.LoadByKid(ctx, val)
	case "name":
		return store.LoadByName(ctx, val)
	}
	return nil, fmt.Errorf(`can't load JWK by %s'`, by)
}

func BadRemoteJwkSetRequestFunc() func(ctx context.Context) *http.Request {
	var i int
	return func(ctx context.Context) *http.Request {
		defer func() {i++}()
		switch i%3 {
		case 0:
			return nil
		case 1:
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:55555/auth/v2/jwks", nil)
			return req
		default:
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8900/auth/not/exist", nil)
			return req
		}
	}

}

func BadRemoteJwkRequestFunc() func(ctx context.Context, kid string) *http.Request {
	var i int
	return func(ctx context.Context, kid string) *http.Request {
		defer func() {i++}()
		switch i%3 {
		case 0:
			return nil
		case 1:
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:55555/auth/v2/jwks/"+kid, nil)
			return req
		default:
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8900/auth/not/exist/"+kid, nil)
			return req
		}
	}
}