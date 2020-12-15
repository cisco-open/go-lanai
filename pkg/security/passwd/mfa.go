package passwd

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"

type MFAEvent int

const (
	_ = iota
	MFAEventOtpCreate
	MFAEventOtpRefresh
	MFAEventVerificationSuccess
	MFAEventVerificationFailure
)

type MFAEventListenerFunc func(event MFAEvent, otp OTP, principal interface{})

/*****************************
	Common Implements
 *****************************/
func broadcastMFAEvent(event MFAEvent, otp OTP, account security.Account, listeners... MFAEventListenerFunc) {
	for _,listener := range listeners {
		listener(event, otp, account)
	}
}
