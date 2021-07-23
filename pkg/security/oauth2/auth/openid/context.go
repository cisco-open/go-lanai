package openid

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

var logger = log.New("OpenID")

const (
	PromptNone       = `none`
	PromptLogin      = `login`
	PromptConsent    = `consent`
	PromptSelectAcct = `select_account`
)

const (
	DisplayPage = `page`
	PromptTouch = `touch`
	//PromptPopup = `popup`
	//PromptWap   = `wap`
)

var (
	ResponseTypes       = utils.NewStringSet("token", "code", "id_token")
	SupportedGrantTypes = utils.NewStringSet(
		oauth2.GrantTypeAuthCode,
		oauth2.GrantTypeImplicit,
		oauth2.GrantTypePassword,
		oauth2.GrantTypeSwitchUser,
		oauth2.GrantTypeSwitchTenant,
	)
	SupportedDisplayMode = utils.NewStringSet(DisplayPage, PromptTouch)
)

