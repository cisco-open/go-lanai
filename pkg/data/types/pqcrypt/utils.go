package pqcrypt

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"fmt"
	"reflect"
	"strings"
)

func newInvalidFormatError(text string, args...interface{}) error {
	return data.NewDataError(data.ErrorCodeOrmMapping, "invalid encrypted data: " + fmt.Sprintf(text, args...))
}

func newEncryptionError(text string, args...interface{}) error {
	return data.NewDataError(data.ErrorCodeOrmMapping, "failed to encrypt data: " + fmt.Sprintf(text, args...))
}

func newDecryptionError(text string, args...interface{}) error {
	return data.NewDataError(data.ErrorCodeOrmMapping, "failed to decrypt data: " + fmt.Sprintf(text, args...))
}

func normalizeKeyID(kid string) string {
	return strings.ToLower(kid)
}

func tryAssign(v interface{}, dest interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = newDecryptionError("recovered: %v", e)
		}
	}()

	// check
	rDest := reflect.ValueOf(dest)
	if rDest.Kind() != reflect.Ptr {
		return newDecryptionError("%T is not assignable", dest)
	}
	rDest = rDest.Elem()
	if !rDest.CanSet() {
		return newDecryptionError("%T is not assignable", dest)
	}

	// assign
	if v == nil {
		rDest.Set(reflect.New(rDest.Type()).Elem())
		return nil
	}

	rv := reflect.ValueOf(v)
	if !rv.Type().AssignableTo(rDest.Type()) {
		return newDecryptionError("decrypted data type mismatch, expect %T, but got %T", rDest.Interface(), v)
	}
	rDest.Set(rv)
	return nil
}