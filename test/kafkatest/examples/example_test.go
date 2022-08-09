package examples

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/kafka"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/kafkatest"
	"fmt"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

const (
	ExampleTopic = "TEST_EVENTS"
	ExampleGroup = "TESTS"
)

/*************************
	Setup
 *************************/

type tsDI struct {
	fx.In
	Binder kafka.Binder
}

type TestMessage struct {
	Int    int
	String string
}

type TestService struct {
	producer kafka.Producer
}

func NewTestService(di tsDI) *TestService {
	svc := &TestService{}
	p, e := di.Binder.Produce(ExampleTopic, kafka.BindingName("test"), kafka.RequireLocalAck())
	if e != nil {
		panic(e)
	}
	svc.producer = p

	s, e := di.Binder.Subscribe(ExampleTopic)
	if e != nil {
		panic(e)
	}
	if e := s.AddHandler(svc.Handle); e != nil {
		panic(e)
	}

	c, e := di.Binder.Consume(ExampleTopic, ExampleGroup)
	if e != nil {
		panic(e)
	}
	if e := c.AddHandler(svc.Handle); e != nil {
		panic(e)
	}
	return svc
}

func (s *TestService) GenerateSomeMessages(ctx context.Context, count int) error {
	for i := 0; i < count; i++ {
		e := s.producer.Send(ctx, &TestMessage{
			Int:    i,
			String: fmt.Sprintf("Message-%d", i),
		}, kafka.WithKey(uuid.New()))
		if e != nil {
			return e
		}
	}
	return nil
}

func (s *TestService) Handle(_ context.Context, _ *kafka.Message) error {
	//noop
	return nil
}

/*************************
	Tests
 *************************/

type testDI struct {
	fx.In
	Service  *TestService
	Recorder kafkatest.MessageRecorder
}

func TestMockedBinder(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		kafkatest.WithMockedBinder(),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(NewTestService),
		),
		test.GomegaSubTest(SubTestExampleProducerRecording(di), "ExampleProducerRecording"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestExampleProducerRecording(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// don't forget to reset recorder
		di.Recorder.Reset()

		var e error
		// DoA something that producing messages
		e = di.Service.GenerateSomeMessages(ctx, 3)
		g.Expect(e).To(Succeed(), "functions using producers shouldn't fail")

		// validate recorded messages
		actual := di.Recorder.Records(ExampleTopic)
		g.Expect(actual).To(HaveLen(3), "recorded messages should have correct length")
		for i, record := range actual {
			g.Expect(record.Payload).To(BeAssignableToTypeOf(&TestMessage{}), "recorded message at [%d] should have correct type", i)
			msg := record.Payload.(*TestMessage)
			g.Expect(msg.Int).To(BeEquivalentTo(i), "recorded message at [%d] should have correct Int field", i)
			g.Expect(msg.String).To(BeEquivalentTo(fmt.Sprintf("Message-%d", i)), "recorded message at [%d] should have correct String field", i)
		}
	}
}


