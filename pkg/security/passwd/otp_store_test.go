package passwd

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/embedded"
    "github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Test Setup
 *************************/

func NewInMemoryOTPStore() OTPStore {
	return inmemOtpStore{}
}

func NewRedisOTPStore(client redis.Client) OTPStore {
	return newRedisOtpStore(client)
}

/*************************
	Test
 *************************/

type OtpDI struct {
	fx.In
	Store OTPStore
}

func TestOTPManagerWithInMemoryOTPStore(t *testing.T) {
	var di OtpDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithFxOptions(
			fx.Provide(NewInMemoryOTPStore),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestVerifySuccessful(&di), "VerifySuccessful"),
		test.GomegaSubTest(SubTestVerifyUnknownOTP(&di), "VerifyUnknownOTP"),
		test.GomegaSubTest(SubTestVerifyWrongPasscode(&di), "VerifyWrongPasscode"),
		test.GomegaSubTest(SubTestRefreshPasscode(&di), "RefreshPasscode"),
		test.GomegaSubTest(SubTestDeleteOTP(&di), "DeleteOTP"),
	)
}

func TestOTPManagerWithRedisOTPStore(t *testing.T) {
    var di OtpDI
    test.RunTest(context.Background(), t,
        apptest.Bootstrap(),
        embedded.WithRedis(),
        apptest.WithModules(redis.Module),
        apptest.WithFxOptions(
            fx.Provide(NewRedisOTPStore),
        ),
        apptest.WithDI(&di),
        test.GomegaSubTest(SubTestVerifySuccessful(&di), "VerifySuccessful"),
        test.GomegaSubTest(SubTestVerifyUnknownOTP(&di), "VerifyUnknownOTP"),
        test.GomegaSubTest(SubTestVerifyWrongPasscode(&di), "VerifyWrongPasscode"),
        test.GomegaSubTest(SubTestRefreshPasscode(&di), "RefreshPasscode"),
        test.GomegaSubTest(SubTestDeleteOTP(&di), "DeleteOTP"),
    )
}

/*************************
	Sub Tests
 *************************/

func SubTestVerifySuccessful(di *OtpDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		manager := newTotpManager(func(manager *totpManager) {
			manager.store = di.Store
            manager.factory = newTotpFactory()
		})
		otp, e := manager.New()
		g.Expect(e).To(Succeed(), "New() should not fail")
        g.Expect(otp.ID()).ToNot(BeEmpty(), "OTP id should be correct")

        otp, e = manager.Get(otp.ID())
        g.Expect(e).To(Succeed(), "get existing OTP should not fail")
        g.Expect(otp).ToNot(BeNil(), "get existing OTP should not return nil")

        matched, moreChance, e := manager.Verify(otp.ID(), otp.Passcode())
        g.Expect(e).To(Succeed(), "Verify() should not fail")
        g.Expect(matched.ID()).To(Equal(otp.ID()), "Verify() should return correct OTP")
        g.Expect(moreChance).To(BeTrue(), "Verify() should return correct 'more chances'")
    }
}

func SubTestVerifyUnknownOTP(di *OtpDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        manager := newTotpManager(func(manager *totpManager) {
            manager.store = di.Store
            manager.factory = newTotpFactory()
        })

        matched, moreChance, e := manager.Verify("unknown", "whatever")
        g.Expect(e).To(HaveOccurred(), "Verify() should fail")
        g.Expect(matched).To(BeNil(), "Verify() should return nil OTP")
        g.Expect(moreChance).To(BeFalse(), "Verify() should return correct 'more chances'")
    }
}

func SubTestVerifyWrongPasscode(di *OtpDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        manager := newTotpManager(func(manager *totpManager) {
            manager.store = di.Store
            manager.factory = newTotpFactory()
            manager.maxVerifyLimit = 2
        })
        otp, e := manager.New()
        g.Expect(e).To(Succeed(), "New() should not fail")
        g.Expect(otp.ID()).ToNot(BeEmpty(), "OTP id should be correct")

        otp, e = manager.Get(otp.ID())
        g.Expect(e).To(Succeed(), "get existing OTP should not fail")
        g.Expect(otp).ToNot(BeNil(), "get existing OTP should not return nil")

        // 1
        matched, moreChance, e := manager.Verify(otp.ID(), "wrong")
        g.Expect(e).To(HaveOccurred(), "Verify() should fail")
        g.Expect(matched.ID()).To(Equal(otp.ID()), "Verify() should return correct OTP")
        g.Expect(moreChance).To(BeTrue(), "Verify() should return correct 'more chances'")

        // 2
        matched, moreChance, e = manager.Verify(otp.ID(), "wrong")
        g.Expect(e).To(HaveOccurred(), "Verify() should fail")
        g.Expect(matched.ID()).To(Equal(otp.ID()), "Verify() should return correct OTP")
        g.Expect(moreChance).To(BeFalse(), "Verify() should return correct 'more chances'")

        // 3
        matched, moreChance, e = manager.Verify(otp.ID(), otp.Passcode())
        g.Expect(e).To(HaveOccurred(), "Verify() should fail")
        g.Expect(matched).To(BeNil(), "Verify() should return nil OTP after reaching verify limit")
        g.Expect(moreChance).To(BeFalse(), "Verify() should return correct 'more chances'")
    }
}

func SubTestRefreshPasscode(di *OtpDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        manager := newTotpManager(func(manager *totpManager) {
            manager.store = di.Store
            manager.factory = newTotpFactory()
            manager.maxRefreshLimit = 2
        })
        otp, e := manager.New()
        g.Expect(e).To(Succeed(), "New() should not fail")
        g.Expect(otp.ID()).ToNot(BeEmpty(), "OTP id should be correct")

        otp, e = manager.Get(otp.ID())
        g.Expect(e).To(Succeed(), "get existing OTP should not fail")
        g.Expect(otp).ToNot(BeNil(), "get existing OTP should not return nil")

        // 0
        matched, moreChance, e := manager.Refresh("unknown")
        g.Expect(e).To(HaveOccurred(), "Refresh() should fail on unknown ID")
        g.Expect(matched).To(BeNil(), "Refresh() return correct OTP on unknown ID")

        // 1
        matched, moreChance, e = manager.Refresh(otp.ID())
        g.Expect(e).To(Succeed(), "Refresh() should not fail")
        g.Expect(matched.ID()).To(Equal(otp.ID()), "Refresh() should return correct OTP")
        g.Expect(moreChance).To(BeTrue(), "Refresh() return correct 'more changes'")

        // 2
        matched, moreChance, e = manager.Refresh(otp.ID())
        g.Expect(e).To(Succeed(), "Refresh() should not fail")
        g.Expect(matched.ID()).To(Equal(otp.ID()), "Refresh() should return correct OTP")
        g.Expect(moreChance).To(BeFalse(), "Refresh() return correct 'more changes'")

        // 3
        matched, moreChance, e = manager.Refresh(otp.ID())
        g.Expect(e).To(HaveOccurred(), "Refresh() should fail after reaches limit")
        g.Expect(matched.ID()).To(Equal(otp.ID()), "Refresh() return correct OTP after reaches limit")
        g.Expect(moreChance).To(BeFalse(), "Refresh() return correct 'more changes'")
    }
}

func SubTestDeleteOTP(di *OtpDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        manager := newTotpManager(func(manager *totpManager) {
            manager.store = di.Store
            manager.factory = newTotpFactory()
            manager.maxRefreshLimit = 2
        })
        otp, e := manager.New()
        g.Expect(e).To(Succeed(), "New() should not fail")
        g.Expect(otp.ID()).ToNot(BeEmpty(), "OTP id should be correct")

        otp, e = manager.Get(otp.ID())
        g.Expect(e).To(Succeed(), "get existing OTP should not fail")
        g.Expect(otp).ToNot(BeNil(), "get existing OTP should not return nil")

        // 1
        e = manager.Delete(otp.ID())
        g.Expect(e).To(Succeed(), "Delete() should not fail")

        otp, e = manager.Get(otp.ID())
        g.Expect(e).To(HaveOccurred(), "get deleted OTP should fail")
        g.Expect(otp).To(BeNil(), "get deleted OTP should return nil")
    }
}
