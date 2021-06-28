package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"time"
)

type PasswordAuthFeature struct {
	accountStore    security.AccountStore
	passwordEncoder PasswordEncoder

	// MFA support
	mfaEnabled        bool
	mfaEventListeners []MFAEventListenerFunc
	otpTTL            time.Duration
	otpVerifyLimit    uint
	otpRefreshLimit   uint
	otpLength         uint
	otpSecretSize     uint
}

// Configure is Standard security.Feature entrypoint
func Configure(ws security.WebSecurity) *PasswordAuthFeature {
	feature := &PasswordAuthFeature{}
	if fm, ok := ws.(security.FeatureModifier); ok {
		return fm.Enable(feature).(*PasswordAuthFeature)
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// New is Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
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

func (f *PasswordAuthFeature) MFAEventListeners(handlers ...MFAEventListenerFunc) *PasswordAuthFeature {
	f.mfaEventListeners = append(f.mfaEventListeners, handlers...)
	return f
}

func (f *PasswordAuthFeature) OtpTTL(ttl time.Duration) *PasswordAuthFeature {
	f.otpTTL = ttl
	return f
}

func (f *PasswordAuthFeature) OtpVerifyLimit(v uint) *PasswordAuthFeature {
	f.otpVerifyLimit = v
	return f
}

func (f *PasswordAuthFeature) OtpRefreshLimit(v uint) *PasswordAuthFeature {
	f.otpRefreshLimit = v
	return f
}

func (f *PasswordAuthFeature) OtpLength(v uint) *PasswordAuthFeature {
	f.otpLength = v
	return f
}

func (f *PasswordAuthFeature) OtpSecretSize(v uint) *PasswordAuthFeature {
	f.otpSecretSize = v
	return f
}
