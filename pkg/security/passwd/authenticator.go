package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
)

const (
	MessageUserNotFound = "Mismatched Username and Password"
	MessageBadCredential = "Mismatched Username and Password"
	MessageOtpNotAvailable = "MFA required but temprorily unavailable"
	MessageAccountStatus = "Inactive Account"
)

/******************************
	security.Authenticator
******************************/
type Authenticator struct {
	accountStore      security.AccountStore
	passwdEncoder     PasswordEncoder
	otpManager        OTPManager
	mfaEventListeners []MFAEventListenerFunc
	checkers 		  []AuthenticationDecisionMaker
}

type AuthenticatorOptionsFunc func(*AuthenticatorOptions)

type AuthenticatorOptions struct {
	AccountStore      security.AccountStore
	PasswordEncoder   PasswordEncoder
	OTPManager        OTPManager
	MFAEventListeners []MFAEventListenerFunc
	Checkers 			  []AuthenticationDecisionMaker
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
		accountStore:      options.AccountStore,
		passwdEncoder:     options.PasswordEncoder,
		otpManager:        options.OTPManager,
		mfaEventListeners: options.MFAEventListeners,
		checkers: 		   options.Checkers,
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

	// pre checks
	ctx := utils.NewMutableContext()
	if err := performChecks(a.checkers, ctx, upp, user, nil); err != nil {
		return nil, a.translate(err)
	}

	// Check password
	if password, ok := user.Credentials().(string);
		!ok || upp.Username != user.Username() || !a.passwdEncoder.Matches(upp.Password, password) {
		return nil, security.NewBadCredentialsError(MessageBadCredential)
	}

	// create authentication
	auth, err := a.CreateSuccessAuthentication(upp, user)
	if err != nil {
		return nil, a.translate(err)
	}

	// post checks
	if err := performChecks(a.checkers, ctx, upp, user, auth); err != nil {
		return nil, a.translate(err)
	}

	return auth, nil
}

// exported for override posibility
func (a *Authenticator) CreateSuccessAuthentication(candidate *UsernamePasswordPair, account security.Account) (security.Authentication, error) {

	permissions := map[string]interface{}{}

	// MFA support
	if candidate.EnforceMFA == MFAModeMust || candidate.EnforceMFA != MFAModeSkip && account.UseMFA() {
		// MFA required
		if a.otpManager == nil {
			return nil, security.NewInternalAuthenticationError(MessageOtpNotAvailable)
		}

		otp, err := a.otpManager.New()
		if err != nil {
			return nil, security.NewInternalAuthenticationError(MessageOtpNotAvailable)
		}
		permissions[SpecialPermissionMFAPending] = true
		permissions[SpecialPermissionOtpId] = otp.ID()

		broadcastMFAEvent(MFAEventOtpCreate, otp, account, a.mfaEventListeners...)
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

	return &auth, nil
}

func (a *Authenticator) translate(err error) error {

	switch {
	case errors.Is(err, security.ErrorTypeSecurity):
		return err
	default:
		return security.NewAccountStatusError(MessageAccountStatus, err)
	}
}


