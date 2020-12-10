package passwd



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
