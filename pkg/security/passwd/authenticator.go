package passwd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"errors"
	"sort"
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
	postProcessors	  []PostAuthenticationProcessor
}

type AuthenticatorOptionsFunc func(*AuthenticatorOptions)

type AuthenticatorOptions struct {
	AccountStore      security.AccountStore
	PasswordEncoder   PasswordEncoder
	OTPManager        OTPManager
	MFAEventListeners []MFAEventListenerFunc
	Checkers          []AuthenticationDecisionMaker
	PostProcessors    []PostAuthenticationProcessor
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
	sort.SliceStable(options.Checkers, func(i,j int) bool {
		return order.OrderedFirstCompare(options.Checkers[i], options.Checkers[j])
	})
	sort.SliceStable(options.PostProcessors, func(i,j int) bool {
		return order.OrderedFirstCompareReverse(options.PostProcessors[i], options.PostProcessors[j])
	})
	return &Authenticator{
		accountStore:      options.AccountStore,
		passwdEncoder:     options.PasswordEncoder,
		otpManager:        options.OTPManager,
		mfaEventListeners: options.MFAEventListeners,
		checkers:          options.Checkers,
		postProcessors:    options.PostProcessors,
	}
}

func (a *Authenticator) Authenticate(ctx context.Context, candidate security.Candidate) (auth security.Authentication, err error) {
	upp, ok := candidate.(*UsernamePasswordPair)
	if !ok {
		return nil, nil
	}

	// schedule post processing
	ctx = utils.MakeMutableContext(ctx)
	var user security.Account
	defer func() {
		auth, err = applyPostAuthenticationProcessors(a.postProcessors, ctx, user, candidate, auth, err)
	}()

	// Search user in the slice of allowed credentials
	user, e := a.accountStore.LoadAccountByUsername(ctx, upp.Username)
	if e != nil {
		err = security.NewUsernameNotFoundError(MessageUserNotFound, e)
		return
	}

	// pre checks
	if e := makeDecision(a.checkers, ctx, upp, user, nil); e != nil {
		err = a.translate(e)
		return
	}

	// Check password
	if password, ok := user.Credentials().(string);
		!ok || upp.Username != user.Username() || !a.passwdEncoder.Matches(upp.Password, password) {

		err = security.NewBadCredentialsError(MessageBadCredential)
		return
	}

	// create authentication
	newAuth, e := a.CreateSuccessAuthentication(upp, user)
	if e != nil {
		err = a.translate(e)
		return
	}

	// post checks
	if e := makeDecision(a.checkers, ctx, upp, user, newAuth); e != nil {
		err = a.translate(e)
		return
	}

	auth = newAuth
	return
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

	details := candidate.DetailsMap
	if details == nil {
		details = map[interface{}]interface{}{}
	}
	auth := usernamePasswordAuthentication{
		Acct:       account,
		Perms:      permissions,
		DetailsMap: details,
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


