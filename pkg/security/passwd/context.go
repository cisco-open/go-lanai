package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"encoding/gob"
	"time"
)

const (
	SpecialPermissionMFAPending = "MFAPending"
	SpecialPermissionOtpId = "OtpId"
)
/******************************
	Serialization
******************************/
func GobRegister() {
	gob.Register((*usernamePasswordAuthentication)(nil))
	gob.Register((*timeBasedOtp)(nil))
	gob.Register(TOTP{})
	gob.Register(time.Time{})
	gob.Register(time.Duration(0))
}

/************************
	security.Candidate
************************/
type MFAMode int
const(
	MFAModeSkip = iota
	MFAModeOptional
	MFAModeMust
)
// UsernamePasswordPair is the supported security.Candidate of this authenticator
type UsernamePasswordPair struct {
	Username string
	Password string
	DetailsMap map[interface{}]interface{}
	EnforceMFA MFAMode
}

// security.Candidate
func (upp *UsernamePasswordPair) Principal() interface{} {
	return upp.Username
}

// security.Candidate
func (upp *UsernamePasswordPair) Credentials() interface{} {
	return upp.Password
}

// security.Candidate
func (upp *UsernamePasswordPair) Details() interface{} {
	return upp.DetailsMap
}

// MFAOtpVerification is the supported security.Candidate for MFA authentication
type MFAOtpVerification struct {
	CurrentAuth UsernamePasswordAuthentication
	OTP string
}

// security.Candidate
func (uop *MFAOtpVerification) Principal() interface{} {
	return uop.CurrentAuth.Principal()
}

// security.Candidate
func (uop *MFAOtpVerification) Credentials() interface{} {
	return uop.OTP
}

// security.Candidate
func (uop *MFAOtpVerification) Details() interface{} {
	return nil
}

// MFAOtpRefresh is the supported security.Candidate for MFA OTP refresh
type MFAOtpRefresh struct {
	CurrentAuth UsernamePasswordAuthentication
}

// security.Candidate
func (uop *MFAOtpRefresh) Principal() interface{} {
	return uop.CurrentAuth.Principal()
}

// security.Candidate
func (uop *MFAOtpRefresh) Credentials() interface{} {
	return uop.CurrentAuth.OTPIdentifier()
}

// security.Candidate
func (uop *MFAOtpRefresh) Details() interface{} {
	return nil
}

/******************************
	security.Authentication
******************************/
// UsernamePasswordAuthentication implements security.Authentication
type UsernamePasswordAuthentication interface {
	security.Authentication
	IsMFAPending() bool
	OTPIdentifier() string
}

// usernamePasswordAuthentication
// Note: all fields should not be used directly. It's exported only because gob only deal with exported field
type usernamePasswordAuthentication struct {
	Acct       security.Account
	Perms      map[string]interface{}
	DetailsMap map[interface{}]interface{}
}

func (auth *usernamePasswordAuthentication) Principal() interface{} {
	return auth.Acct
}

func (auth *usernamePasswordAuthentication) Permissions() map[string]interface{} {
	return auth.Perms
}

func (auth *usernamePasswordAuthentication) State() security.AuthenticationState {
	switch {
	case auth.IsMFAPending():
		return security.StatePrincipalKnown
	default:
		return security.StateAuthenticated
	}
}

func (auth *usernamePasswordAuthentication) Details() interface{} {
	return auth.Acct
}

func (auth *usernamePasswordAuthentication) IsMFAPending() bool {
	_, ok := auth.Permissions()[SpecialPermissionOtpId].(string)
	return ok
}

func (auth *usernamePasswordAuthentication) OTPIdentifier() string {
	v, ok := auth.Permissions()[SpecialPermissionOtpId].(string)
	if ok {
		return v
	}
	return ""
}

func IsSamePrincipal(username string, currentAuth security.Authentication) bool {
	if currentAuth == nil || currentAuth.State() < security.StatePrincipalKnown {
		return false
	}

	if account, ok := currentAuth.Principal().(security.Account); ok && username == account.Username() {
		return true
	} else if principal, ok := currentAuth.Principal().(string); ok && username == principal {
		return true
	}
	return false
}