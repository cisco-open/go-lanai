package datacrypto

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
	ErrUnsupportedVersion = fmt.Errorf("unsupported version of encrypted data format")
	ErrInvalidFormat = fmt.Errorf("encrypted data is invalid")

	v1TextPrefix = fmt.Sprintf("%s%s", V1, v1Separator)
	javaTypePattern, _ = regexp.Compile(`^[[:alpha:]][[:alnum:]]*(\.[[:alpha:]][[:alnum:]]*)+`)
)

/*************************
	Data Carrier
 *************************/

type EncryptedData struct {
	Ver  Version     `json:"v"`
	UUID uuid.UUID   `json:"kid"`
	Alg  Algorithm   `json:"alg"`
	Raw  interface{} `json:"d"`
}

// UnmarshalTextV1 deserialize V1 format of text
func (d *EncryptedData) UnmarshalTextV1(text []byte) error {
	str := string(text)
	if !isV1Format(str) {
		return ErrUnsupportedVersion
	}

	split := strings.SplitN(str, v1Separator, 4)
	if len(split) < 4 {
		return ErrInvalidFormat
	}

	var ver Version
	if e := unmarshalText(split[0], &ver); e != nil {
		return fmt.Errorf("%v: %v", ErrInvalidFormat, e)
	}

	kid, e := uuid.Parse(split[1])
	if e != nil {
		return fmt.Errorf("%v: %v", ErrInvalidFormat, e)
	}

	var alg Algorithm
	if e := unmarshalText(split[2], &alg); e != nil {
		return fmt.Errorf("%v: %v", ErrInvalidFormat, e)
	}

	raw, e := unmarshalV1RawPayload(alg, split[3])
	if e != nil {
		return e
	}

	*d = EncryptedData{
		Ver:  ver,
		UUID: kid,
		Alg:  alg,
		Raw:  raw,
	}
	return nil
}

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
			return fmt.Errorf("%s: %s", ErrInvalidFormat, e)
		}
		// find first non-string, also check if string element is a Java type expr
		for i, elem := range s {
			switch v := elem.(type) {
			case string:
				if i % 2 == 0 && !javaTypePattern.Match([]byte(v)) {
					return fmt.Errorf("%s: invalid Java type", ErrInvalidFormat)
				}
			default:
				d.value = elem
			}
		}
	case '{':
		var v map[string]interface{}
		if e := json.Unmarshal(data, &v); e != nil {
			return fmt.Errorf("%s: %s", ErrInvalidFormat, e)
		}
		d.value = v
	default:
		return ErrInvalidFormat
	}
	return nil
}

/*************************
	Data Types Ext
 *************************/

// UnmarshalText implements encoding.TextUnmarshaler with V1 support
func (d *EncryptedMap) UnmarshalText(data []byte) error {
	// first try V1
	switch e := (&d.EncryptedData).UnmarshalTextV1(data); {
	case e == nil:
		return nil
	case e != ErrUnsupportedVersion:
		return e
	}

	// try JSON format
	return json.Unmarshal(data, &d.EncryptedData)
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
//	- Encrypted: a string
func unmarshalV1RawPayload(alg Algorithm, text string) (interface{}, error) {
	switch alg {
	case AlgPlain:
		raw := rawPlainDataV1{}
		if e := json.Unmarshal([]byte(text), &raw); e != nil {
			return nil, e
		}
		return raw.value, nil
	case AlgVault:
		fallthrough
	default:
		return text, nil
	}
}