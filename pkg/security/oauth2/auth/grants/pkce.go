package grants

import (
	"crypto"
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	PKCEChallengeMethodPlain  PKCECodeChallengeMethod = "plain"
	PKCEChallengeMethodSHA256 PKCECodeChallengeMethod = "S256"
)

type PKCECodeChallengeMethod string

func (m *PKCECodeChallengeMethod) UnmarshalText(text []byte) error {
	str := string(text)
	switch {
	case string(PKCEChallengeMethodPlain) == strings.ToLower(str):
		*m = PKCEChallengeMethodPlain
	case string(PKCEChallengeMethodSHA256) == strings.ToUpper(str):
		*m = PKCEChallengeMethodSHA256
	case len(text) == 0:
		*m = PKCEChallengeMethodPlain
	default:
		return fmt.Errorf("invalid code challenge method")
	}
	return nil
}

func parseCodeChallengeMethod(str string) (ret PKCECodeChallengeMethod, err error) {
	err = ret.UnmarshalText([]byte(str))
	return
}

// https://datatracker.ietf.org/doc/html/rfc7636#section-4.6
func verifyPKCE(toVerify string, challenge string, method PKCECodeChallengeMethod) (ret bool) {
	var encoded string
	switch method {
	case PKCEChallengeMethodPlain:
		encoded = toVerify
	case PKCEChallengeMethodSHA256:
		hash := crypto.SHA256.New()
		if _, e := hash.Write([]byte(toVerify)); e != nil {
			return
		}
		encoded = base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
	default:
		return
	}
	return encoded == challenge
}