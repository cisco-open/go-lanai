package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"time"
)

type PasswordAuthFeature struct {
	accountStore     security.AccountStore
	passwordEncoder  PasswordEncoder

	// MFA support
	mfaEnabled        bool
	mfaEventListeners []MFAEventListenerFunc
	otpTTL            time.Duration
	otpVerifyLimit    uint
	otpRefreshLimit   uint
}

// Standard security.Feature entrypoint
func Configure(ws security.WebSecurity) *PasswordAuthFeature {
	feature := &PasswordAuthFeature{}
	if fm, ok := ws.(security.FeatureModifier); ok {
		return fm.Enable(feature).(*PasswordAuthFeature)
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *PasswordAuthFeature {
	return &PasswordAuthFeature{}
}

func (f *PasswordAuthFeature) Identifier() security.FeatureIdentifier {
	return PasswordAuthenticatorFeatureId
}

func (f *PasswordAuthFeature) AccountStore(as security.AccountStore) *PasswordAuthFeature {
	f.accountStore = as
	return f
}

func (f *PasswordAuthFeature) PasswordEncoder(pe PasswordEncoder) *PasswordAuthFeature {
	f.passwordEncoder = pe
	return f
}

func (f *PasswordAuthFeature) MFA(enabled bool) *PasswordAuthFeature {
	f.mfaEnabled = enabled
	return f
}

func (f *PasswordAuthFeature) MFAEventListeners(handlers...MFAEventListenerFunc) *PasswordAuthFeature {
	f.mfaEventListeners = append(f.mfaEventListeners, handlers...)
	return f
}

func (f *PasswordAuthFeature) OtpTTL(ttl time.Duration) *PasswordAuthFeature {
	f.otpTTL = ttl
	return f
}

func (f *PasswordAuthFeature) OtpVerifyLimit(count uint) *PasswordAuthFeature {
	f.otpVerifyLimit = count
	return f
}

func (f *PasswordAuthFeature) OtpRefreshLimit(count uint) *PasswordAuthFeature {
	f.otpRefreshLimit = count
	return f
}