package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

const (
	MessageInvalidPasscode = "Invalid Passcode"
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
	switch _, more, err := a.otpStore.Verify(id, verify.OTP); {
	case err != nil && more:
		return nil, security.NewBadCredentialsError(MessageInvalidPasscode, err)
	case err != nil:
		msg := fmt.Errorf("%s: %s", MessageInvalidPasscode, err.Error())
		return nil, security.NewBadCredentialsError(msg, err)
	}

	// TODO post passcode check

	auth, err := a.CreateSuccessAuthentication(verify, user)
	return auth, nil
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

func (a *MfaVerifyAuthenticator) notifyMfaEvent(event MFAEvent, otp OTP, account security.Account) {
	for _,listener := range a.mfaEventListeners {
		listener(event, otp, account)
	}
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

