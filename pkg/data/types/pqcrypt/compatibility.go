package pqcrypt

import (
	"encoding"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"regexp"
	"strings"
)

const (
	v1Separator = ":"
)

var (
	v1TextPrefix = fmt.Sprintf("%d%s", V1, v1Separator)
	javaTypePattern, _ = regexp.Compile(`^[[:alpha:]][[:alnum:]]*(\.[[:alpha:]][[:alnum:]]*)+`)
)

/*************************
	Parsing
 *************************/

func ParseEncryptedRaw(text string) (*EncryptedRaw, error) {
	v := &EncryptedRaw{}

	// first try V1
	switch e := v.UnmarshalTextV1([]byte(text)); {
	case e == nil:
		return v, nil
	case e != ErrUnsupportedVersion:
		return nil, e
	}

	// try JSON format
	if e := json.Unmarshal([]byte(text), v); e != nil {
		return nil, newInvalidFormatError("invalid V2 format - %v", e)
	}
	return v, nil
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

	kid, e := uuid.Parse(split[1])
	if e != nil {
		return newInvalidFormatError("invalid Key ID")
	}

	var alg Algorithm
	if e := unmarshalText(split[2], &alg); e != nil {
		return newInvalidFormatError("unsupported algorithm")
	}

	raw, e := unmarshalV1RawPayload(alg, split[3])
	if e != nil {
		return e
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
	Data Carrier
 *************************/

type rawPlainDataV1 struct {
	value interface{}
}

// UnmarshalJSON implements json.Unmarshaler with V1 support
func (d *rawPlainDataV1) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 0 {
		d.value = map[string]interface{}{}
		return nil
	}
	switch data[0] {
	case '[':
		var s []interface{}
		if e := json.Unmarshal(data, &s); e != nil {
			return e
		}
		// find first non-string, also check if string element is a Java type expr
		for i, elem := range s {
			switch v := elem.(type) {
			case string:
				if i % 2 == 0 && !javaTypePattern.Match([]byte(v)) {
					return fmt.Errorf("malformed legacy Java type")
				}
			default:
				d.value = elem
			}
		}
	case '{':
		var v map[string]interface{}
		if e := json.Unmarshal(data, &v); e != nil {
			return e
		}
		d.value = v
	default:
		return fmt.Errorf("only JSON array or object are supported")
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

// unmarshalV1RawPayload V1 (Java) format of raw payload and convert it to object or string
// V1 format could be
// 	- Plain: a (T extends Map<String, String>) serialized by Jackson with `As.WRAPPER_ARRAY` option
//	- EncryptedRaw: a string
func unmarshalV1RawPayload(alg Algorithm, text string) (interface{}, error) {
	switch alg {
	case AlgPlain:
		raw := rawPlainDataV1{}
		if e := json.Unmarshal([]byte(text), &raw); e != nil {
			return nil, newInvalidFormatError("raw data JSON parsing error - %v", e)
		}
		return raw.value, nil
	case AlgVault:
		fallthrough
	default:
		return text, nil
	}
}