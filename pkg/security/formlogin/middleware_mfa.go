package formlogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"github.com/gin-gonic/gin"
)

var (

)

type MfaAuthenticationMiddleware struct {
	authenticator  security.Authenticator
	successHandler security.AuthenticationSuccessHandler
	otpParam       string
}

type MfaMWOptionsFunc func(*MfaMWOptions)

type MfaMWOptions struct {
	Authenticator  security.Authenticator
	SuccessHandler security.AuthenticationSuccessHandler
	OtpParam       string
}

func NewMfaAuthenticationMiddleware(optionFuncs ...MfaMWOptionsFunc) *MfaAuthenticationMiddleware {
	options := MfaMWOptions{}
	for _, optFunc := range optionFuncs {
		if optFunc != nil {
			optFunc(&options)
		}
	}
	return &MfaAuthenticationMiddleware{
		authenticator:  options.Authenticator,
		successHandler: options.SuccessHandler,
		otpParam:       options.OtpParam,
	}
}

func (mw *MfaAuthenticationMiddleware) OtpVerifyHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		otp := ctx.PostFormArray(mw.otpParam)
		if len(otp) == 0 {
			otp = []string{""}
		}

		before, err := mw.currentAuth(ctx);
		if err != nil {
			mw.handleError(ctx, err, nil)
			return
		}

		candidate := passwd.MFAOtpVerification{
			CurrentAuth: before,
			OTP:         otp[0],
			DetailsMap:  map[interface{}]interface{}{},
		}

		// authenticate
		auth, err := mw.authenticator.Authenticate(&candidate)
		if err != nil {
			mw.handleError(ctx, err, &candidate)
			return
		}
		mw.handleSuccess(ctx, before, auth)
	}
}

func (mw *MfaAuthenticationMiddleware) OtpRefreshHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		before, err := mw.currentAuth(ctx);
		if err != nil {
			mw.handleError(ctx, err, nil)
			return
		}
		candidate := passwd.MFAOtpRefresh{
			CurrentAuth: before,
			DetailsMap:  map[interface{}]interface{}{},
		}

		// authenticate
		auth, err := mw.authenticator.Authenticate(&candidate)
		if err != nil {
			mw.handleError(ctx, err, &candidate)
			return
		}
		mw.handleSuccess(ctx, before, auth)
	}
}

func (mw *MfaAuthenticationMiddleware) EndpointHandlerFunc() gin.HandlerFunc {
	return notFoundHandlerFunc
}

func (mw *MfaAuthenticationMiddleware) currentAuth(ctx *gin.Context) (passwd.UsernamePasswordAuthentication, error) {
	if currentAuth, ok := security.Get(ctx).(passwd.UsernamePasswordAuthentication); !ok || !currentAuth.IsMFAPending() {
		return nil, security.NewAccessDeniedError("MFA is not in progess")
	} else {
		return currentAuth, nil
	}
}

func (mw *MfaAuthenticationMiddleware) handleSuccess(c *gin.Context, before, new security.Authentication) {
	if new != nil {
		c.Set(gin.AuthUserKey, new.Principal())
		c.Set(security.ContextKeySecurity, new)
	}
	mw.successHandler.HandleAuthenticationSuccess(c, c.Request, c.Writer, before, new)
	if c.Writer.Written() {
		c.Abort()
	}
}

func (mw *MfaAuthenticationMiddleware) handleError(c *gin.Context, err error, candidate security.Candidate) {
	if mw.shouldClear(err) {
		security.Clear(c)
	}
	_ = c.Error(err)
	c.Abort()
}

func (mw *MfaAuthenticationMiddleware) shouldClear(err error) bool {
	switch coder, ok := err.(security.ErrorCoder); ok {
	case coder.Code() == security.ErrorCodeCredentialsExpired:
		return true
	case coder.Code() == security.ErrorCodeMaxAttemptsReached:
		return true
	}
	return false
}
