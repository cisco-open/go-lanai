package oauth2

import "encoding/gob"

func init() {
	gob.Register((*authentication)(nil))
	gob.Register((*OAuth2Request)(nil))
}
