package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

/*************************
	Test
 *************************/


func TestFilterResource(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(5 * time.Minute),
		apptest.WithModules(opa.Module),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestFilterByTenantID(di), "TestFilterByTenantID"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestFilterByTenantID(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// member admin
		ctx = sectest.ContextWithSecurity(ctx, memberAdminOptions())
		// member admin - can read
		result, e := FilterResource(ctx, "poc", opa.OpRead, func(res *Resource) {})
		g.Expect(e).To(Succeed())
		fmt.Println(result)
	}
}


