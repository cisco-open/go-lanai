package oauth2

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"encoding/gob"
)

var logger = log.New("OAuth2")

func init() {
	gob.Register((*authentication)(nil))
	gob.Register((*OAuth2Request)(nil))
	gob.Register((*OAuth2Error)(nil))
}
