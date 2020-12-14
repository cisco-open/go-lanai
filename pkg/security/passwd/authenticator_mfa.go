package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

const (
	MessageInvalidPasscode = "Invalid Passcode"
	MessageCannotRefresh = "Unable to refresh passcode"
	MessageInvalidAccountStatus = "Issue with current account status"
	MessageMaxAttemptsReached = "You have reached the maximum allowed passcode verification attempts"
	MessageMaxRefreshAttemptsReached = "You have reached the maximum allowed passcode refresh attempts"
)

/********************************
	MfaVerifyAuthenticator
*********************************/
type MfaVerifyAuthenticator struct {
	accountStore      security.AccountStore
	otpStore          OTPStore
	mfaEventListeners []MFAEventListenerFunc
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
		otpStore:          options.OTPStore,
		mfaEventListeners: options.MFAEventListeners,
	}
}

func (a *MfaVerifyAuthenticator) Authenticate(candidate security.Candidate) (security.Authentication, error) {
	verify, ok := candidate.(*MFAOtpVerification)
	if !ok {
		return nil, nil
	}

	// check if OTP verification should be performed
	user, err := checkCurrentAuth(verify.CurrentAuth, a.accountStore)
	if err != nil {
		return nil, err
	}

	// Check OTP
	id := verify.CurrentAuth.OTPIdentifier()
	switch otp, more, err := a.otpStore.Verify(id, verify.OTP); {
	case err != nil && more:
		notifyMfaEvent(a.mfaEventListeners, MFAEventVerificationFailure, otp, user)
		return nil, security.NewBadCredentialsError(MessageInvalidPasscode, err)
	case err != nil:
		notifyMfaEvent(a.mfaEventListeners, MFAEventVerificationFailure, otp, user)
		msg := fmt.Errorf("%s: %s", MessageInvalidPasscode, err.Error())
		return nil, security.NewBadCredentialsError(msg, err)
	default:
		notifyMfaEvent(a.mfaEventListeners, MFAEventVerificationSuccess, otp, user)
	}

	// TODO post passcode check

	return a.CreateSuccessAuthentication(verify, user)
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

/********************************
	MfaVerifyAuthenticator
*********************************/
type MfaRefreshAuthenticator struct {
	accountStore      security.AccountStore
	otpStore          OTPStore
	mfaEventListeners []MFAEventListenerFunc
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
		otpStore:          options.OTPStore,
		mfaEventListeners: options.MFAEventListeners,
	}
}

func (a *MfaRefreshAuthenticator) Authenticate(candidate security.Candidate) (security.Authentication, error) {
	refresh, ok := candidate.(*MFAOtpRefresh)
	if !ok {
		return nil, nil
	}

	// check if OTP refresh should be performed
	user, err := checkCurrentAuth(refresh.CurrentAuth, a.accountStore)
	if err != nil {
		return nil, err
	}

	// Refresh OTP
	id := refresh.CurrentAuth.OTPIdentifier()
	switch otp, more, err := a.otpStore.Refresh(id); {
	case err != nil && more:
		return nil, security.NewBadCredentialsError(MessageCannotRefresh, err)
	case err != nil:
		msg := fmt.Errorf("%s: %s", MessageCannotRefresh, err.Error())
		return nil, security.NewBadCredentialsError(msg, err)
	default:
		notifyMfaEvent(a.mfaEventListeners, MFAEventOtpRefresh, otp, user)
	}

	// TODO post passcode refresh

	return a.CreateSuccessAuthentication(refresh, user)
}

// exported for override posibility
func (a *MfaRefreshAuthenticator) CreateSuccessAuthentication(candidate *MFAOtpRefresh, account security.Account) (security.Authentication, error) {
	// TODO chance for other components to add details
	return candidate.CurrentAuth, nil
}

/************************
	Helpers
 ************************/
func checkCurrentAuth(currentAuth UsernamePasswordAuthentication, accountStore security.AccountStore) (security.Account, error) {
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

	user, err := accountStore.LoadAccountByUsername(username)
	if err != nil {
		return nil, security.NewUsernameNotFoundError(MessageInvalidAccountStatus, err)
	}
	// TODO check account status
	return user, nil
}

func notifyMfaEvent(listeners []MFAEventListenerFunc, event MFAEvent, otp OTP, account security.Account) {
	for _,listener := range listeners {
		listener(event, otp, account)
	}
}

