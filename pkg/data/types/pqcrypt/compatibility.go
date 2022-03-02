package pqcrypt

import (
	"encoding"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"regexp"
	"strconv"
	"strings"
)

const (
	v1Separator = ":"
)

var (
	v1TextPrefix = fmt.Sprintf("%d%s", V1, v1Separator)
	javaTypePattern, _ = regexp.Compile(`^[[:alpha:]][[:alnum:]]*(\.[[:alpha:]][[:alnum:]]*)+`)
	jsonNull = `null`
)

/*************************
	Parsing
 *************************/

func ParseEncryptedRaw(text string) (ret *EncryptedRaw, err error) {
	ret = &EncryptedRaw{}

	// first try V1
	//nolint:errorlint // special error is global var
	switch e := ret.UnmarshalTextV1([]byte(text)); {
	case e == nil:
		return
	case e != ErrUnsupportedVersion:
		return nil, e
	}

	// try JSON format
	if e := json.Unmarshal([]byte(text), ret); e != nil {
		return nil, newInvalidFormatError("invalid V2 format - %v", e)
	}
	return
}

// UnmarshalTextV1 deserialize V1 format of text
func (d *EncryptedRaw) UnmarshalTextV1(text []byte) error {
	str := string(text)
	if !isV1Format(str) {
		return ErrUnsupportedVersion
	}

	split := strings.SplitN(str, v1Separator, 4)
	if len(split) < 4 {
		return newInvalidFormatError("not V1 format")
	}

	var ver Version
	if e := unmarshalText(split[0], &ver); e != nil {
		return newInvalidFormatError("unsupported version")
	}

	kid:= split[1]
	if _, e := uuid.Parse(kid); e != nil {
		return newInvalidFormatError("invalid Key ID")
	}

	var alg Algorithm
	if e := unmarshalText(split[2], &alg); e != nil {
		return newInvalidFormatError("unsupported algorithm")
	}

	var raw json.RawMessage
	switch alg {
	case AlgPlain:
		raw = json.RawMessage(split[3])
	case AlgVault:
		raw = json.RawMessage(strconv.Quote(split[3]))
	}
	if !json.Valid(raw) {
		return newInvalidFormatError("unsupported raw data")
	}

	*d = EncryptedRaw{
		Ver:   ver,
		KeyID: kid,
		Alg:   alg,
		Raw:   raw,
	}
	return nil
}

/*************************
	V1 Plain Data
 *************************/

type v1DecryptedData json.RawMessage

// UnmarshalJSON implements json.Unmarshaler with V1 support
// V1 (Java) format of unencrypted payload could be
// 	- a (T extends Map<String, String>) serialized by Jackson with `As.WRAPPER_ARRAY` option (JSON Array)
//  - a (T extends Map>String, String>) serialized by Jackson without `As.WRAPPER_ARRAY` option (JSON Object
func (d *v1DecryptedData) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 0 {
		*d = nil
		return nil
	}
	switch data[0] {
	case '[':
		var s []json.RawMessage
		if e := json.Unmarshal(data, &s); e != nil {
			return e
		}
		// find first non-string, also check if string element is a Java type expr
		var v json.RawMessage
		switch len(s) {
		case 1:
			v = s[0]
		case 2:
			str := ""
			if e := json.Unmarshal(s[0], &str); e != nil || !javaTypePattern.Match([]byte(str)) {
				return ErrInvalidV1Format
			}
			v = s[1]
		default:
			return ErrInvalidV1Format
		}
		*d = v1DecryptedData(v)
	case '{':
		*d = data
	default:
		return ErrInvalidV1Format
	}
	return nil
}

/*************************
	helpers
 *************************/

func isV1Format(text string) bool {
	return strings.HasPrefix(text, v1TextPrefix)
}

func unmarshalText(data string, v encoding.TextUnmarshaler) error {
	return v.UnmarshalText([]byte(data))
}

// extractV1DecryptedPayload decode V1 (Java) format of unencrypted payload and convert it to object
// V1 format could be
// 	- a (T extends Map<String, String>) serialized by Jackson with `As.WRAPPER_ARRAY` option (JSON Array)
//  - a (T extends Map>String, String>) serialized by Jackson without `As.WRAPPER_ARRAY` option (JSON Object
//  - a JSON "null"
//  - empty data
func extractV1DecryptedPayload(data []byte) (json.RawMessage, error) {
	if len(data) == 0 || len(data) == 4 && string(data) == jsonNull {
		// json null or nil/empty is considered nil
		return json.RawMessage(jsonNull), nil
	}

	raw := v1DecryptedData{}
	if e := json.Unmarshal(data, &raw); e != nil {
		return nil, newInvalidFormatError("unencrypted data JSON parsing error - %v", e)
	}
	return json.RawMessage(raw), nil
}