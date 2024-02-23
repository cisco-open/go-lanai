package passwd_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	. "cto-github.cisco.com/NFV-BU/go-lanai/test/utils/gomega"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
	"time"
)

const (
	TestUser         = "test-user"
	TestUserPassword = "ShowMeTheMoney"
	TestPermission   = "ALLOW_TEST"

	LockoutTimeout         = time.Second
	LockoutFailureInterval = time.Second
	PasswordMaxAge = 120 * time.Second
	PasswordWarningDuration = 60 * time.Second
)

/*************************
	Test
 *************************/

type AuthDI struct {
	fx.In
}

func TestAuthenticator(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestSuccessWithoutMFA(), "SuccessWithoutMFA"),
		test.GomegaSubTest(SubTestUserNotFound(), "UserNotFound"),
		test.GomegaSubTest(SubTestWrongPassword(), "WrongPassword"),
		test.GomegaSubTest(SubTestDisabledAccount(), "DisabledAccount"),
		test.GomegaSubTest(SubTestAutoLockout(), "AutoLockout"),
		test.GomegaSubTest(SubTestAutoLockoutException(), "AutoLockoutException"),
		test.GomegaSubTest(SubTestExpiredCredentials(), "ExpiredCredentials"),
		test.GomegaSubTest(SubTestExpiringCredentialsWarning(), "ExpiringCredentialsWarning"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestSuccessWithoutMFA() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		acct := NewAccount(TestUser, TestUserPassword)
		authenticator := NewTestAuthenticator(ctx, g, acct)
		auth, e := authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser, TestUserPassword))
		g.Expect(e).To(Succeed(), "authentication should not fail")
		g.Expect(auth.State()).To(Equal(security.StateAuthenticated), "auth state should be correct")
	}
}

func SubTestUserNotFound() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        acct := NewAccount(TestUser, TestUserPassword)
        authenticator := NewTestAuthenticator(ctx, g, acct)
        _, e := authenticator.Authenticate(ctx, NewCredentialsCandidate("another-user","whatever"))
        g.Expect(e).To(HaveOccurred(), "authentication should fail")
		g.Expect(e).To(IsError(security.NewUsernameNotFoundError("")), "error should be correct")
    }
}

func SubTestWrongPassword() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        acct := NewAccount(TestUser, TestUserPassword)
        authenticator := NewTestAuthenticator(ctx, g, acct)
        _, e := authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser,"whatever"))
        g.Expect(e).To(HaveOccurred(), "authentication should fail")
		g.Expect(e).To(IsError(security.NewBadCredentialsError("")), "error should be correct")
        g.Expect(acct.SerialFailedAttempts()).To(Equal(1), "failed attempt should be recorded")
		g.Expect(acct.LoginFailures()).To(HaveLen(1), "login failures should be recorded")
    }
}

func SubTestDisabledAccount() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		acct := NewAccount(TestUser, TestUserPassword)
		authenticator := NewTestAuthenticator(ctx, g, acct)
		acct.AcctDetails.Disabled = true
		_, e := authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser,TestUserPassword))
		g.Expect(e).To(HaveOccurred(), "authentication should fail")
		g.Expect(e).To(IsError(security.NewAccountStatusError("")), "error should be correct")
		// note: disabled account check should happen before password validation
		g.Expect(acct.SerialFailedAttempts()).To(Equal(0), "failed attempt should not be recorded")
		g.Expect(acct.LoginFailures()).To(HaveLen(0), "login failures should not be recorded")
	}
}

func SubTestAutoLockout() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        var e error
        var auth security.Authentication
        acct := NewAccount(TestUser, TestUserPassword)
        authenticator := NewTestAuthenticator(ctx, g, acct)
        // 1
        _, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser,"whatever"))
        g.Expect(e).To(HaveOccurred(), "authentication should fail")
		g.Expect(e).To(IsError(security.NewBadCredentialsError("")), "error should be correct")
		g.Expect(acct.LoginFailures()).To(HaveLen(1), "login failures should be recorded")

        // 2
        _, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser,"whatever"))
        g.Expect(e).To(HaveOccurred(), "authentication should fail")
		g.Expect(e).To(IsError(security.NewAccountStatusError("")), "error should be correct")
        g.Expect(acct.Locked()).To(BeTrue(), "account should be locked out")

        // after lockout
        _, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser,TestUserPassword))
        g.Expect(e).To(HaveOccurred(), "authentication should fail")
		g.Expect(e).To(IsError(security.NewAccountStatusError("")), "error should be correct after lockout")

        // move back lockout time and test auto unlock
        acct.AcctDetails.LockoutTime = time.Now().Add(-LockoutTimeout)
        auth, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser,TestUserPassword))
        g.Expect(e).To(Succeed(), "authentication should not fail")
        g.Expect(auth.State()).To(Equal(security.StateAuthenticated), "auth state should be correct")
		g.Expect(acct.LoginFailures()).To(BeEmpty(), "login failures should be reset")
		g.Expect(acct.SerialFailedAttempts()).To(Equal(0), "failed attempt count should be reset")
        g.Expect(acct.Locked()).To(BeFalse(), "account should be unlocked")
    }
}

func SubTestAutoLockoutException() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		acct := NewAccount(TestUser, TestUserPassword)
		authenticator := NewTestAuthenticator(ctx, g, acct)
		// 1
		_, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser,"whatever"))
		g.Expect(e).To(HaveOccurred(), "authentication should fail")
		g.Expect(e).To(IsError(security.NewBadCredentialsError("")), "error should be correct")
		g.Expect(acct.LoginFailures()).To(HaveLen(1), "login failures should be recorded")

		// 2 move last failure time outside of "serial failure" range, then attempt #1 wouldn't be counted for auto lockout
		acct.AcctDetails.LoginFailures[0] = time.Now().Add(-LockoutFailureInterval)
		_, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser,"whatever"))
		g.Expect(e).To(HaveOccurred(), "authentication should fail")
		g.Expect(e).To(IsError(security.NewBadCredentialsError("")), "error should be correct")
		g.Expect(acct.Locked()).To(BeFalse(), "account should not be locked out")
		g.Expect(acct.LoginFailures()).To(HaveLen(2), "login failures records should be adjusted")
		g.Expect(acct.SerialFailedAttempts()).To(Equal(2), "failed attempt count should be correct")
	}
}

func SubTestExpiredCredentials() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var auth security.Authentication
		acct := NewAccount(TestUser, TestUserPassword)
		authenticator := NewTestAuthenticator(ctx, g, acct)

		acct.AcctDetails.PwdChangedTime = time.Now().Add(-PasswordMaxAge)
		// 1
		auth, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser, TestUserPassword))
		g.Expect(e).To(Succeed(), "authentication should not fail")
		g.Expect(auth.State()).To(Equal(security.StateAuthenticated), "auth state should be correct")
		g.Expect(acct.GracefulAuthCount()).To(Equal(1), "graceful auth count should be increased")

		// 2 should fail due to expired credentials
		_, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser,TestUserPassword))
		g.Expect(e).To(HaveOccurred(), "authentication should fail")
		g.Expect(e).To(IsError(security.NewCredentialsExpiredError("")), "error should be correct")
		g.Expect(acct.Locked()).To(BeFalse(), "account should not be locked out")

		// 3 mark password updated and try again with incorrect password
		acct.AcctDetails.PwdChangedTime = time.Now()
		// In case of password changed, we don't require "lockout period" to be waited before auto-unlock account
		//acct.AcctDetails.LockoutTime = time.Now().Add(-LockoutTimeout)
		auth, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser, "whatever"))
		g.Expect(e).To(HaveOccurred(), "authentication should fail")
		g.Expect(e).To(IsError(security.NewBadCredentialsError("")), "error should be correct")
		g.Expect(acct.Locked()).To(BeFalse(), "account should be unlocked")

		// 4
		auth, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser, TestUserPassword))
		g.Expect(e).To(Succeed(), "authentication should not fail")
		g.Expect(auth.State()).To(Equal(security.StateAuthenticated), "auth state should be correct")
		g.Expect(acct.SerialFailedAttempts()).To(Equal(0), "failed attempt count should be reset")
	}
}

func SubTestExpiringCredentialsWarning() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var auth security.Authentication
		acct := NewAccount(TestUser, TestUserPassword)
		authenticator := NewTestAuthenticator(ctx, g, acct)
		// 1
		auth, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser, TestUserPassword))
		g.Expect(e).To(Succeed(), "authentication should not fail")
		g.Expect(auth.State()).To(Equal(security.StateAuthenticated), "auth state should be correct")
		g.Expect(acct.GracefulAuthCount()).To(Equal(0), "graceful auth count should not be increased")
		g.Expect(auth.Details()).ToNot(HaveKey(security.DetailsKeyAuthWarning), "auth details should not contains warning")

		// 2
		acct.AcctDetails.PwdChangedTime = time.Now().Add(-PasswordWarningDuration)
		auth, e = authenticator.Authenticate(ctx, NewCredentialsCandidate(TestUser, TestUserPassword))
		g.Expect(e).To(Succeed(), "authentication should not fail")
		g.Expect(auth.State()).To(Equal(security.StateAuthenticated), "auth state should be correct")
		g.Expect(acct.GracefulAuthCount()).To(Equal(0), "graceful auth count should not be increased")
		g.Expect(auth.Details()).To(HaveKey(security.DetailsKeyAuthWarning), "auth details should contains warning")
	}
}

/*************************
	Helpers
 *************************/

func NewTestAuthenticator(ctx context.Context, g *gomega.WithT, accts ...security.Account) security.Authenticator {
	props := make([]*sectest.MockedAccountProperties, len(accts))
	overrideLookup := map[string]security.Account{}
	for i := range accts {
		props[i] = &sectest.MockedAccountProperties{
			UserId:   accts[i].ID().(string),
			Username: accts[i].Username(),
		}
		overrideLookup[accts[i].Username()] = accts[i]
	}
	store := sectest.NewMockedAccountStore(props, func(acct security.Account) security.Account {
		if override, ok := overrideLookup[acct.Username()]; ok && override != nil {
			return override
		}
		return acct
	})
	authn, e := passwd.NewAuthenticatorBuilder(passwd.New().
		AccountStore(store).MFA(true),
	).Build(ctx)
	g.Expect(e).To(Succeed(), "building authenticator should not fail")
	return authn
}

func NewCredentialsCandidate(username, password string) security.Candidate {
	return &passwd.UsernamePasswordPair{
		Username:   username,
		Password:   password,
		EnforceMFA: passwd.MFAModeOptional,
	}
}

func NewAccount(username, password string, opts ...func(acct *security.DefaultAccount)) *security.DefaultAccount {
	acct := &security.DefaultAccount{
		AcctDetails: security.AcctDetails{
            ID:                        utils.RandomString(16),
            Type:                      security.AccountTypeDefault,
            Username:                  username,
            Credentials:               password,
            Permissions:               []string{TestPermission},
            PwdChangedTime:            time.Now(),
            PolicyName:                "test-password-rule",
        },
		AcctLockingRule: security.AcctLockingRule{
			Name:             "test-locking-rule",
			Enabled:          true,
			LockoutDuration:  LockoutTimeout,
			FailuresLimit:    2,
			FailuresInterval: LockoutFailureInterval,
		},
		AcctPasswordPolicy: security.AcctPasswordPolicy{
            Name:                "test-password-rule",
            Enabled:             true,
            MaxAge:              PasswordMaxAge,
            ExpiryWarningPeriod: PasswordWarningDuration,
            GracefulAuthLimit:   1,
        },
		AcctMetadata: security.AcctMetadata{},
	}
    for _, fn := range opts {
        fn(acct)
    }
    return acct
}
