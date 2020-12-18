package passwd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
)

const (
	MessageInvalidPasscode = "Bad Verification Code"
	MessagePasscodeExpired = "Verification Code Expired"
	MessageCannotRefresh = "Unable to Refresh"
	MessageMaxAttemptsReached = "No More Verification Attempts Allowed"
	MessageMaxRefreshAttemptsReached = "No More Resend Attempts Allowed"
	MessageInvalidAccountStatus = "Issue with current account status"
)

// For error translation
var (
	errorBadCredentials     = security.NewBadCredentialsError("bad creds")
	errorCredentialsExpired = security.NewCredentialsExpiredError("cred exp")
	errorMaxAttemptsReached = security.NewMaxAttemptsReachedError("max attempts")
	errorAccountStatus      = security.NewAccountStatusError("acct status")
)

/********************************
	MfaVerifyAuthenticator
*********************************/
type MfaVerifyAuthenticator struct {
	accountStore      security.AccountStore
	otpStore          OTPManager
	mfaEventListeners []MFAEventListenerFunc
	checkers 		  []AuthenticationDecisionMaker
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
	}
}

func (a *MfaVerifyAuthenticator) Authenticate(candidate security.Candidate) (security.Authentication, error) {
	verify, ok := candidate.(*MFAOtpVerification)
	if !ok {
		return nil, nil
	}

	// check if OTP verification should be performed
	ctx := utils.NewMutableContext()
	user, err := checkCurrentAuth(ctx, verify.CurrentAuth, a.accountStore)
	if err != nil {
		return nil, err
	}

	// pre checks
	if err := performChecks(a.checkers, ctx, verify, user, nil); err != nil {
		return nil, a.translate(err, true)
	}

	// Check OTP
	id := verify.CurrentAuth.OTPIdentifier()
	switch otp, more, err := a.otpStore.Verify(id, verify.OTP); {
	case err != nil:
		broadcastMFAEvent(MFAEventVerificationFailure, otp, user, a.mfaEventListeners...)
		return nil, a.translate(err, more)
	default:
		broadcastMFAEvent(MFAEventVerificationSuccess, otp, user, a.mfaEventListeners...)
	}

	auth, err := a.CreateSuccessAuthentication(verify, user)
	if err != nil {
		return auth, a.translate(err, true)
	}

	// post checks
	if err := performChecks(a.checkers, ctx, verify, user, auth); err != nil {
		return nil, a.translate(err, true)
	}
	return auth, err
}

// exported for override posibility
func (a *MfaVerifyAuthenticator) CreateSuccessAuthentication(candidate *MFAOtpVerification, account security.Account) (security.Authentication, error) {

	permissions := map[string]interface{}{}
	for _,p := range account.Permissions() {
		permissions[p] = true
	}

	details, ok := candidate.CurrentAuth.Details().(map[interface{}]interface{})
	if !ok {
		details := map[interface{}]interface{}{}
		details["Literal"] = candidate.CurrentAuth.Details()
	}

	auth := usernamePasswordAuthentication{
		Acct:       account,
		Perms:      permissions,
		DetailsMap: details,
	}
	// TODO chance for other components to add details
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
	checkers 			  []AuthenticationDecisionMaker
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
		checkers: 		   options.Checkers,
	}
}

func (a *MfaRefreshAuthenticator) Authenticate(candidate security.Candidate) (security.Authentication, error) {
	refresh, ok := candidate.(*MFAOtpRefresh)
	if !ok {
		return nil, nil
	}

	// check if OTP refresh should be performed
	ctx := utils.NewMutableContext()
	user, err := checkCurrentAuth(ctx, refresh.CurrentAuth, a.accountStore)
	if err != nil {
		return nil, err
	}

	// pre checks
	if err := performChecks(a.checkers, ctx, refresh, user, nil); err != nil {
		return nil, a.translate(err, true)
	}

	// Refresh OTP
	id := refresh.CurrentAuth.OTPIdentifier()
	switch otp, more, err := a.otpStore.Refresh(id); {
	case err != nil:
		return nil, a.translate(err, more)
	default:
		broadcastMFAEvent(MFAEventOtpRefresh, otp, user, a.mfaEventListeners...)
	}

	auth, err := a.CreateSuccessAuthentication(refresh, user)
	if err != nil {
		return auth, a.translate(err, true)
	}

	// post checks
	if err := performChecks(a.checkers, ctx, refresh, user, auth); err != nil {
		return nil, a.translate(err, true)
	}
	return auth, err
}

// exported for override posibility
func (a *MfaRefreshAuthenticator) CreateSuccessAuthentication(candidate *MFAOtpRefresh, account security.Account) (security.Authentication, error) {
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

	var username string
	switch currentAuth.Principal().(type) {
	case string:
		username = currentAuth.Principal().(string)
	case security.Account:
		username = currentAuth.Principal().(security.Account).Username()
	}

	user, err := accountStore.LoadAccountByUsername(ctx, username)
	if err != nil {
		return nil, security.NewUsernameNotFoundError(MessageInvalidAccountStatus, err)
	}

	return user, nil
}

