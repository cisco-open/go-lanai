package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
	"sort"
	"time"
)

var (
	PasswordAuthenticatorFeatureId = security.PriorityFeatureId("passwdAuth", security.FeatureOrderAuthenticator)
)

type PasswordAuthConfigurer struct {
	accountStore security.AccountStore
	passwordEncoder PasswordEncoder
	redisClient redis.Client
}

func newPasswordAuthConfigurer(store security.AccountStore, encoder PasswordEncoder, redisClient redis.Client) *PasswordAuthConfigurer {
	return &PasswordAuthConfigurer {
		accountStore:    store,
		passwordEncoder: encoder,
		redisClient:     redisClient,
	}
}

func (pac *PasswordAuthConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Verify
	if err := pac.validate(feature.(*PasswordAuthFeature), ws); err != nil {
		return err
	}
	f := feature.(*PasswordAuthFeature)

	// options
	defaultOpts := pac.defaultOptions(f)
	var mfaOpts AuthenticatorOptionsFunc
	if f.mfaEnabled {
		mfaOpts = pac.mfaOptions(f)
	}

	// username passowrd authenticator
	auth := NewAuthenticator(defaultOpts, mfaOpts)
	ws.Authenticator().(*security.CompositeAuthenticator).Add(auth)

	// MFA
	if f.mfaEnabled {
		mfaVerify := NewMFAVerifyAuthenticator(defaultOpts, mfaOpts)
		ws.Authenticator().(*security.CompositeAuthenticator).Add(mfaVerify)

		mfaRefresh := NewMFARefreshAuthenticator(defaultOpts, mfaOpts)
		ws.Authenticator().(*security.CompositeAuthenticator).Add(mfaRefresh)
	}
	return nil
}

func (pac *PasswordAuthConfigurer) validate(f *PasswordAuthFeature, ws security.WebSecurity) error {

	if _,ok := ws.Authenticator().(*security.CompositeAuthenticator); !ok {
		return fmt.Errorf("unable to add password authenticator to %T", ws.Authenticator())
	}

	if f.accountStore == nil && pac.accountStore == nil {
		return fmt.Errorf("unable to create password authenticator: account accountStore is not set")
	}
	return nil
}

func (pac *PasswordAuthConfigurer) defaultOptions(f *PasswordAuthFeature) AuthenticatorOptionsFunc {
	if f.accountStore == nil {
		f.accountStore = pac.accountStore
	}

	if f.passwordEncoder == nil {
		f.passwordEncoder = pac.passwordEncoder
	}

	// TODO maybe customizeble via Feature
	acctStatusChecker := NewAccountStatusChecker(f.accountStore)
	passwordChecker := NewPasswordPolicyChecker(f.accountStore)

	return func(opts *AuthenticatorOptions) {
		opts.AccountStore = f.accountStore
		if f.passwordEncoder != nil {
			opts.PasswordEncoder = f.passwordEncoder
		}
		opts.Checkers = []AuthenticationDecisionMaker{
			PreCheck(acctStatusChecker),
			FinalCheck(passwordChecker),
		}
	}
}

func (pac *PasswordAuthConfigurer) mfaOptions(f *PasswordAuthFeature) AuthenticatorOptionsFunc {
	if !f.mfaEnabled {
		return func(*AuthenticatorOptions) {}
	}

	if f.otpTTL <= 0 {
		f.otpTTL = 10 * time.Minute
	}

	if f.otpVerifyLimit <= 0 {
		f.otpVerifyLimit = 3
	}

	if f.otpRefreshLimit <= 0 {
		f.otpRefreshLimit = 3
	}

	otpManager := newTotpManager(func(s *totpManager) {
		s.ttl = f.otpTTL
		s.maxVerifyLimit = f.otpVerifyLimit
		s.maxRefreshLimit = f.otpRefreshLimit
		if pac.redisClient != nil {
			s.store = newRedisOtpStore(pac.redisClient)
		}
	})

	// TODO maybe customizeble via Feature
	acctStatusChecker := NewAccountStatusChecker(f.accountStore)
	passwordChecker := NewPasswordPolicyChecker(f.accountStore)

	return func(opts *AuthenticatorOptions) {
		opts.OTPManager = otpManager
		sort.SliceStable(f.mfaEventListeners, func(i,j int) bool {
			return order.OrderedFirstCompare(f.mfaEventListeners[i], f.mfaEventListeners[j])
		})
		opts.MFAEventListeners = f.mfaEventListeners
		opts.Checkers = []AuthenticationDecisionMaker{
			PreCheck(acctStatusChecker),
			FinalCheck(passwordChecker),
		}
	}
}
