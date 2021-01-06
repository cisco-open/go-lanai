package oauth2

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type valueConverterFunc func(v interface{}) (reflect.Value, error)

/** DefaultAccessToken **/
var accessTokenIgnoredDetails = utils.NewStringSet(
	JsonFieldAccessTokenValue, JsonFieldTokenType, JsonFieldScopes, JsonFieldExpiryTime, JsonFieldExpiresIn)

// json.Marshaler
func (t *DefaultAccessToken) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}
	for k,v := range t.details {
		data[k] = v
	}
	data[JsonFieldAccessTokenValue] = t.value
	data[JsonFieldTokenType] = t.tokenType
	data[JsonFieldScopes] = t.scopes
	data[JsonFieldExpiryTime] = t.expiryTime.Format(utils.ISO8601Seconds)
	data[JsonFieldExpiresIn] = int(t.expiryTime.Sub(time.Now()).Seconds())

	return json.Marshal(data)
}

// json.Unmarshaler
func (t *DefaultAccessToken) UnmarshalJSON(data []byte) error {
	parsed := map[string]interface{}{}

	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}

	if err := extractField(parsed, JsonFieldAccessTokenValue, true, reflect.ValueOf(&t.value), anyToString); err != nil {
		return err
	}

	if err := extractField(parsed, JsonFieldTokenType, true, reflect.ValueOf(&t.tokenType), anyToTokenType); err != nil {
		return err
	}

	if err := extractField(parsed, JsonFieldScopes, true, reflect.ValueOf(&t.scopes), sliceToStringSet); err != nil {
		return err
	}

	// default to parse expiry time from JsonFieldExpiryTime field, fall back to JsonFieldExpiresIn
	if err := extractField(parsed, JsonFieldExpiryTime, true, reflect.ValueOf(&t.expiryTime), expiryToTime); err != nil {
		if err := extractField(parsed, JsonFieldExpiresIn, true, reflect.ValueOf(&t.expiryTime), expireInToTime); err != nil {
			return err
		}
	}

	// put the rest of fields to details
	for k, v := range parsed {
		if !accessTokenIgnoredDetails.Has(k) {
			t.details[k] = v
		}
	}
	return nil
}

func extractField(data map[string]interface{}, field string, required bool, dest reflect.Value, converter valueConverterFunc) error {
	v, ok := data[field]
	switch {
	case !ok && required:
		return fmt.Errorf("cannot find required field [%s]", field)
	case !ok:
		return nil
	}

	value, err := converter(v)
	if err != nil {
		return fmt.Errorf("cannot parse field [%s]: %s", field, err.Error())
	}

	if !dest.CanSet() {
		dest = dest.Elem()
	}

	dest.Set(value)
	return nil
}

func anyToString(v interface{}) (reflect.Value, error) {
	_, ok := v.(string)
	if !ok {
		return reflect.Value{}, fmt.Errorf("invalid field type. expected string")
	}
	return reflect.ValueOf(v), nil
}

func anyToTokenType(v interface{}) (reflect.Value, error) {
	s, ok := v.(string)
	if !ok {
		return reflect.Value{}, fmt.Errorf("invalid field type. expected string")
	}
	return reflect.ValueOf(TokenType(s)), nil
}

func sliceToStringSet(v interface{}) (reflect.Value, error) {
	slice, ok := v.([]interface{})
	if !ok {
		return reflect.Value{}, fmt.Errorf("invalid field type. expected array")
	}

	scopes := utils.NewStringSet()
	for _, s := range slice {
		if str, ok := s.(string); !ok {
			return reflect.Value{}, fmt.Errorf("invalid field type. expected array")
		} else {
			scopes.Add(str)
		}
	}
	return reflect.ValueOf(scopes), nil
}

func expiryToTime(v interface{}) (reflect.Value, error) {
	str, ok := v.(string)
	if !ok {
		return reflect.Value{}, fmt.Errorf("invalid field type. expected ISO8601 formatted string")
	}

	if time := utils.ParseTimeISO8601(str); !time.IsZero() {
		return reflect.ValueOf(time), nil
	} else if time := utils.ParseTime(utils.ISO8601Milliseconds, str); !time.IsZero() {
		return reflect.ValueOf(time), nil
	}

	return reflect.Value{}, fmt.Errorf("invalid field format. expected ISO8601 formatted string")
}

func expireInToTime(v interface{}) (reflect.Value, error) {
	secs, ok := v.(int64)
	if !ok {
		return reflect.Value{}, fmt.Errorf("invalid field type. expected integer")
	}

	time := time.Now().Add(time.Duration(secs) * time.Second)
	return reflect.ValueOf(time), nil
}
