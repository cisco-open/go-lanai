package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"encoding/gob"
)

const (
	SpecialPermissionMFAPending = "MFAPending"
	SpecialPermissionOtpId = "OtpId"
	DetailsKeyUseMFA = "UseMFA"
)

/************************
	security.Candidate
************************/
// UsernamePasswordPair is the supported security.Candidate of this authenticator
type UsernamePasswordPair struct {
	Username string
	Password string
	DetailsMap map[interface{}]interface{}
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

// MFAUsernameOtpPair is the supported security.Candidate for MFA authentication
type MFAUsernameOtpPair struct {
	Username string
	OTP string
}

// security.Candidate
func (uop *MFAUsernameOtpPair) Principal() interface{} {
	return uop.Username
}

// security.Candidate
func (uop *MFAUsernameOtpPair) Credentials() interface{} {
	return uop.OTP
}

// security.Candidate
func (uop *MFAUsernameOtpPair) Details() interface{} {
	return nil
}

// MFAOtpRefresh is the supported security.Candidate for MFA OTP refresh
type MFAOtpRefresh struct {
	Username string
}

// security.Candidate
func (uop *MFAOtpRefresh) Principal() interface{} {
	return uop.Username
}

// security.Candidate
func (uop *MFAOtpRefresh) Credentials() interface{} {
	return nil
}

// security.Candidate
func (uop *MFAOtpRefresh) Details() interface{} {
	return nil
}

/******************************
	Serialization
******************************/
func GobRegister() {
	gob.Register((*usernamePasswordAuthentication)(nil))
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
	account     security.Account
	permissions map[string]interface{}
	details map[interface{}]interface{}
}

func (auth *usernamePasswordAuthentication) Principal() interface{} {
	return auth.account
}

func (auth *usernamePasswordAuthentication) Permissions() map[string]interface{} {
	return auth.permissions
}

func (auth *usernamePasswordAuthentication) Authenticated() bool {
	return true
}

func (auth *usernamePasswordAuthentication) Details() interface{} {
	return auth.account
}

func (auth *usernamePasswordAuthentication) IsMFAPending() bool {
	return security.HasPermissions(auth, SpecialPermissionMFAPending)
}

func (auth *usernamePasswordAuthentication) OTPIdentifier() string {
	v, ok := auth.Permissions()[SpecialPermissionOtpId].(string)
	if ok {
		return v
	}
	return ""
}

func IsSamePrincipal(username string, currentAuth security.Authentication) bool {
	if currentAuth == nil || !currentAuth.Authenticated() {
		return false
	}

	if account, ok := currentAuth.Principal().(security.Account); ok && username == account.Username() {
		return true
	} else if principal, ok := currentAuth.Principal().(string); ok && username == principal {
		return true
	}
	return false
}