package util

import (
	"errors"
	"reflect"
)

func ListContains(list []string, needle string) bool {
	for _, required := range list {
		if needle == required {
			return true
		}
	}
	return false
}

func GetInterfaceType(val interface{}) string {
	t := reflect.TypeOf(val)
	var res string
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		res += "*"
	}
	return res + t.Name()
}

func args(values ...interface{}) []interface{} {
	return values
}

func increment(val int) int {
	return val + 1
}

func templateLog(message ...interface{}) string {
	logger.Infof("%v", message)
	return ""
}

func derefBoolPtr(ptr *bool) (bool, error) {
	if ptr == nil {
		return false, errors.New("pointer is nil")
	}
	return *ptr, nil
}
