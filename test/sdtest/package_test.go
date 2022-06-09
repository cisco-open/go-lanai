package sdtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"embed"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"strconv"
	"testing"
)

var (
	ExpectedServices = map[string]int{
		"service1": 2,
		"service2": 1,
	}
)

//go:embed testdata/*
var testFS embed.FS

type testDI struct {
	fx.In
	Client discovery.Client
}

func TestWithMockedSD(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithMockedSD(LoadDefinition(testFS, "testdata/services.yml")),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestDiscovery(di), "TestDiscovery"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestDiscovery(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Client).To(BeAssignableToTypeOf(&ClientMock{}), "discovery client should be *ClientMock")
		for svc, count := range ExpectedServices {
			instancer, e := di.Client.Instancer(svc)
			g.Expect(e).To(Succeed(), "getting Instancer of %s shouldn't fail", svc)
			insts, e := instancer.Instances(nil)
			g.Expect(e).To(Succeed(), "Instancer.Instances of %s shouldn't fail", svc)
			g.Expect(insts).To(HaveLen(count), "Instancer.Instances should return correct number of instances")
			for i, inst := range insts {
				expectedId := fmt.Sprintf("%s-inst-%d", svc, i)
				expectedAddr := fmt.Sprintf("192.168.0.10%d", i)
				g.Expect(inst.ID).To(Equal(expectedId), "ID should be correct of instance %d", i)
				g.Expect(inst.Address).To(Equal(expectedAddr), "Address should be correct of instance %d", i)
				g.Expect(inst.Port).To(Equal(9000 + i), "Port should be correct of instance %d", i)
				g.Expect(inst.Health).To(Equal(discovery.HealthPassing), "Health should be correct of instance %d", i)
				g.Expect(inst.Tags).To(ContainElements("mocked", "inst" + strconv.Itoa(i)), "Tags should be correct of instance %d", i)
				g.Expect(inst.Meta).To(HaveKeyWithValue("index", strconv.Itoa(i)), "Meta should be correct of instance %d", i)
			}
		}
	}
}