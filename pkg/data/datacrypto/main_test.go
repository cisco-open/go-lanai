package datacrypto

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Test Cases
 *************************/

// Note: Encrypt and Decrypt are covered in map_test.go

func TestCreateKey(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(Module),
		apptest.WithFxOptions(
			fx.Provide(newMockedEncryptor),
		),
		test.GomegaSubTest(SubTestCreateKey(uuid.New(), true), "CreateKeySuccess"),
		test.GomegaSubTest(SubTestCreateKey(uuid.UUID{}, false), "CreateKeyFail"),
	)
}

func TestNoopCreateKey(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(Module),
		test.GomegaSubTest(SubTestCreateKey(uuid.New(), true), "CreateKeyWithValidKey"),
		test.GomegaSubTest(SubTestCreateKey(uuid.UUID{}, true), "CreateKeyWithInvalidKey"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestCreateKey(uuid uuid.UUID, expectSuccess bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := CreateKey(ctx, uuid)
		if expectSuccess {
			g.Expect(e).To(Succeed(), "CreateKey should success")
		} else {
			g.Expect(e).To(Not(Succeed()), "CreateKey should fail")
		}
	}
}
