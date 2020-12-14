package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
)

const (
	MessageUserNotFound = "Mismatched Username and Password"
	MessageBadCredential = "Mismatched Username and Password"
	MessageOtpNotAvailable = "MFA required but temprorily unavailable"
)

/******************************
	security.Authenticator
******************************/
type Authenticator struct {
	accountStore      security.AccountStore
	passwdEncoder     PasswordEncoder
	otpStore          OTPStore
	mfaEventListeners []MFAEventListenerFunc
}

type AuthenticatorOptionsFunc func(*AuthenticatorOptions)

type AuthenticatorOptions struct {
	AccountStore      security.AccountStore
	PasswordEncoder   PasswordEncoder
	OTPStore          OTPStore
	MFAEventListeners []MFAEventListenerFunc
}

func NewAuthenticator(optionFuncs...AuthenticatorOptionsFunc) *Authenticator {
	options := AuthenticatorOptions {
		PasswordEncoder: NewNoopPasswordEncoder(),
		MFAEventListeners: []MFAEventListenerFunc{},
	}
	for _,optFunc := range optionFuncs {
		if optFunc != nil {
			optFunc(&options)
		}
	}
	return &Authenticator{
		accountStore: options.AccountStore,
		passwdEncoder:     options.PasswordEncoder,
		otpStore:          options.OTPStore,
		mfaEventListeners: options.MFAEventListeners,
	}
}

func (a *Authenticator) Authenticate(candidate security.Candidate) (security.Authentication, error) {
	upp, ok := candidate.(*UsernamePasswordPair)
	if !ok {
		return nil, nil
	}

	// Search user in the slice of allowed credentials
	user, err := a.accountStore.LoadAccountByUsername(upp.Username)
	if err != nil {
		return nil, security.NewUsernameNotFoundError(MessageUserNotFound, err)
	}

	// TODO check account status

	// Check password
	if upp.Username != user.Username() || !a.passwdEncoder.Matches(upp.Password, user.Password()) {
		return nil, security.NewBadCredentialsError(MessageBadCredential)
	}

	// TODO post password check

	auth, err := a.CreateSuccessAuthentication(upp, user)
	return auth, nil
}

// exported for override posibility
func (a *Authenticator) CreateSuccessAuthentication(candidate *UsernamePasswordPair, account security.Account) (security.Authentication, error) {

	permissions := map[string]interface{}{}

	// MFA support
	if candidate.EnforceMFA == MFAModeMust || candidate.EnforceMFA != MFAModeSkip && account.UseMFA() {
		// MFA required
		if a.otpStore == nil {
			return nil, security.NewInternalAuthenticationError(MessageOtpNotAvailable)
		}

		otp, err := a.otpStore.New()
		if err != nil {
			return nil, security.NewInternalAuthenticationError(MessageOtpNotAvailable)
		}
		a.notifyMfaEvent(MFAEventOtpCreate, otp, account)
		permissions[SpecialPermissionMFAPending] = true
		permissions[SpecialPermissionOtpId] = otp.ID()
	} else {
		// MFA skipped
		for _,p := range account.Permissions() {
			permissions[p] = true
		}
	}

	auth := usernamePasswordAuthentication{
		Acct:       account,
		Perms:      permissions,
		DetailsMap: candidate.DetailsMap,
	}
	// TODO chance for other components to add details
	return &auth, nil
}

func (a *Authenticator) notifyMfaEvent(event MFAEvent, otp OTP, account security.Account) {
	for _,listener := range a.mfaEventListeners {
		listener(event, otp, account)
	}
}

