package passwd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"time"
)

/********************************
	MfaVerifyAuthenticator
*********************************/

type MfaVerifyAuthenticator struct {
	accountStore      security.AccountStore
	otpStore          OTPManager
	mfaEventListeners []MFAEventListenerFunc
	checkers 		  []AuthenticationDecisionMaker
	postProcessors	  []PostAuthenticationProcessor
}

func NewMFAVerifyAuthenticator(optionFuncs...AuthenticatorOptionsFunc) *MfaVerifyAuthenticator {
	options := AuthenticatorOptions {
		MFAEventListeners: []MFAEventListenerFunc{},
	}
	for _,optFunc := range optionFuncs {
		optFunc(&options)
	}
	return &MfaVerifyAuthenticator{
		accountStore:      options.AccountStore,
		otpStore:          options.OTPManager,
		mfaEventListeners: options.MFAEventListeners,
		checkers: 		   options.Checkers,
		postProcessors:    options.PostProcessors,
	}
}

func (a *MfaVerifyAuthenticator) Authenticate(ctx context.Context, candidate security.Candidate) (auth security.Authentication, err error) {
	verify, ok := candidate.(*MFAOtpVerification)
	if !ok {
		return nil, nil
	}

	// schedule post processing
	ctx = utils.MakeMutableContext(ctx) //nolint:contextcheck
	var user security.Account
	defer func() {
		auth, err = applyPostAuthenticationProcessors(a.postProcessors, ctx, user, candidate, auth, err)
	}()

	// check if OTP verification should be performed
	user, err = checkCurrentAuth(ctx, verify.CurrentAuth, a.accountStore)
	if err != nil {
		return
	}

	// pre checks
	if e := makeDecision(a.checkers, ctx, verify, user, nil); e != nil {
		err = a.translate(e, true)
		return
	}

	// Check OTP
	id := verify.CurrentAuth.OTPIdentifier()
	switch otp, more, e := a.otpStore.Verify(id, verify.OTP); {
	case e != nil:
		broadcastMFAEvent(MFAEventVerificationFailure, otp, user, a.mfaEventListeners...)
		err = a.translate(e, more)
		return
	default:
		broadcastMFAEvent(MFAEventVerificationSuccess, otp, user, a.mfaEventListeners...)
	}

	newAuth, e := a.CreateSuccessAuthentication(verify, user)
	if e != nil {
		err = a.translate(e, true)
		return
	}

	// post checks
	if e := makeDecision(a.checkers, ctx, verify, user, newAuth); e != nil {
		err = e
		return
	}
	auth = newAuth
	return
}

// CreateSuccessAuthentication exported for override posibility
func (a *MfaVerifyAuthenticator) CreateSuccessAuthentication(candidate *MFAOtpVerification, account security.Account) (security.Authentication, error) {

	permissions := map[string]interface{}{}
	for _,p := range account.Permissions() {
		permissions[p] = true
	}

	details, ok := candidate.CurrentAuth.Details().(map[string]interface{})
	if details == nil || !ok {
		details = map[string]interface{}{}
		if candidate.CurrentAuth.Details() != nil {
			details["Literal"] = candidate.CurrentAuth.Details()
		}
	}
	details[security.DetailsKeyAuthTime] = time.Now().UTC()

	auth := usernamePasswordAuthentication{
		Acct:       account,
		Perms:      permissions,
		DetailsMap: details,
	}
	return &auth, nil
}

func (a *MfaVerifyAuthenticator) translate(err error, more bool) error {
	if more {
		return security.NewBadCredentialsError(MessageInvalidPasscode, err)
	}

	switch {
	case errors.Is(err, errorCredentialsExpired):
		return security.NewCredentialsExpiredError(MessagePasscodeExpired, err)
	case errors.Is(err, errorMaxAttemptsReached):
		return security.NewMaxAttemptsReachedError(MessageMaxAttemptsReached, err)
	default:
		return security.NewMaxAttemptsReachedError(MessageInvalidPasscode, err)
	}
}

/********************************
	MfaVerifyAuthenticator
*********************************/

type MfaRefreshAuthenticator struct {
	accountStore      security.AccountStore
	otpStore          OTPManager
	mfaEventListeners []MFAEventListenerFunc
	checkers          []AuthenticationDecisionMaker
	postProcessors    []PostAuthenticationProcessor
}

func NewMFARefreshAuthenticator(optionFuncs...AuthenticatorOptionsFunc) *MfaRefreshAuthenticator {
	options := AuthenticatorOptions {
		MFAEventListeners: []MFAEventListenerFunc{},
	}
	for _,optFunc := range optionFuncs {
		optFunc(&options)
	}
	return &MfaRefreshAuthenticator{
		accountStore:      options.AccountStore,
		otpStore:          options.OTPManager,
		mfaEventListeners: options.MFAEventListeners,
		checkers:          options.Checkers,
		postProcessors:    options.PostProcessors,
	}
}

func (a *MfaRefreshAuthenticator) Authenticate(ctx context.Context, candidate security.Candidate) (auth security.Authentication, err error) {
	refresh, ok := candidate.(*MFAOtpRefresh)
	if !ok {
		return nil, nil
	}

	// schedule post processing
	ctx = utils.MakeMutableContext(ctx) //nolint:contextcheck
	var user security.Account
	defer func() {
		auth, err = applyPostAuthenticationProcessors(a.postProcessors, ctx, user, candidate, auth, err)
	}()

	// check if OTP refresh should be performed
	user, err = checkCurrentAuth(ctx, refresh.CurrentAuth, a.accountStore)
	if err != nil {
		return
	}

	// pre checks
	if e := makeDecision(a.checkers, ctx, refresh, user, nil); e != nil {
		err = a.translate(e, true)
		return
	}

	// Refresh OTP
	id := refresh.CurrentAuth.OTPIdentifier()
	switch otp, more, e := a.otpStore.Refresh(id); {
	case e != nil:
		err = a.translate(e, more)
		return
	default:
		broadcastMFAEvent(MFAEventOtpRefresh, otp, user, a.mfaEventListeners...)
	}

	newAuth, e := a.CreateSuccessAuthentication(refresh, user)
	if e != nil {
		err = a.translate(e, true)
		return
	}

	// post checks
	if e := makeDecision(a.checkers, ctx, refresh, user, newAuth); e != nil {
		err = e
		return
	}

	auth = newAuth
	return
}

// CreateSuccessAuthentication exported for override possibility
func (a *MfaRefreshAuthenticator) CreateSuccessAuthentication(candidate *MFAOtpRefresh, _ security.Account) (security.Authentication, error) {
	return candidate.CurrentAuth, nil
}

func (a *MfaRefreshAuthenticator) translate(err error, more bool) error {
	if more {
		return security.NewBadCredentialsError(MessageCannotRefresh, err)
	}

	switch {
	case errors.Is(err, errorCredentialsExpired):
		return security.NewCredentialsExpiredError(MessagePasscodeExpired, err)
	case errors.Is(err, errorMaxAttemptsReached):
		return security.NewMaxAttemptsReachedError(MessageMaxRefreshAttemptsReached, err)
	default:
		return security.NewMaxAttemptsReachedError(MessageCannotRefresh, err)
	}
}

/************************
	Helpers
 ************************/
func checkCurrentAuth(ctx context.Context, currentAuth UsernamePasswordAuthentication, accountStore security.AccountStore) (security.Account, error) {
	if currentAuth == nil {
		return nil, security.NewUsernameNotFoundError(MessageInvalidAccountStatus)
	}

	user, err := accountStore.LoadAccountByUsername(ctx, currentAuth.Username())
	if err != nil {
		return nil, security.NewUsernameNotFoundError(MessageInvalidAccountStatus, err)
	}

	return user, nil
}

