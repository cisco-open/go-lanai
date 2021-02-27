package auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"encoding/gob"
)

var logger = log.New("OAuth2AuthServer")

func init() {
	gob.Register((*AuthorizeRequest)(nil))
	gob.Register((*TokenRequest)(nil))
}