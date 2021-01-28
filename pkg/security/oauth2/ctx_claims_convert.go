package oauth2

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"reflect"
	"time"
)

var (
	timeType    = reflect.TypeOf(time.Time{})
	int64Type   = reflect.TypeOf(int64(0))
	float64Type = reflect.TypeOf(float64(0))
	float32Type = reflect.TypeOf(float32(0))
	sSliceType = reflect.TypeOf([]string{})
	iSliceType = reflect.TypeOf([]interface{}{})
	sSetType = reflect.TypeOf(utils.NewStringSet())
	iSetType = reflect.TypeOf(utils.NewSet())
	anyType     = reflect.TypeOf(interface{}(0))
)
// some conversions
func claimMarshalConvert(v reflect.Value) (reflect.Value, error) {
	t := v.Type()
	switch {
	case  timeType.AssignableTo(t):
		return timeToTimestamp(v)
	case  float64Type.AssignableTo(t):
		fallthrough
	case  float32Type.AssignableTo(t):
		return v.Convert(int64Type), nil
	default:
		return v, nil
	}
}

func claimUnmarshalConvert(v reflect.Value, fieldType reflect.Type) (reflect.Value, error) {
	switch {
	// special target types
	case  timeType.AssignableTo(fieldType):
		return timestampToTime(v)
	case  sSetType.AssignableTo(fieldType):
		return toStringSet(v)
	case  iSetType.AssignableTo(fieldType):
		return toSet(v)

	// special source types
	case v.Type().AssignableTo(float32Type):
		fallthrough
	case v.Type().AssignableTo(float64Type):
		return v.Convert(int64Type), nil

	// convertable and assignable
	case v.Type().AssignableTo(fieldType):
		return v, nil
	case v.Type().ConvertibleTo(fieldType):
		return v.Convert(fieldType), nil
	default:
		return v, nil
	}
}

func timestampToTime(v reflect.Value) (reflect.Value, error) {
	switch {
	case v.Type().ConvertibleTo(int64Type):
		timestamp := v.Convert(int64Type).Interface().(int64)
		return reflect.ValueOf(time.Unix(timestamp, 0)), nil
	case v.Type().ConvertibleTo(timeType):
		return v.Convert(timeType), nil
	default:
		return reflect.Value{}, fmt.Errorf("type %T cannot be converted to time.Time", v.Interface())
	}
}

func timeToTimestamp(v reflect.Value) (reflect.Value, error) {
	switch {
	case v.Type().ConvertibleTo(timeType):
		time := v.Convert(timeType).Interface().(time.Time)
		return reflect.ValueOf(time.Unix()), nil
	default:
		return reflect.Value{}, fmt.Errorf("type %T cannot be converted to timestamp", v.Interface())
	}
}

func toStringSet(v reflect.Value) (reflect.Value, error) {
	switch {
	case v.Type().ConvertibleTo(sSliceType):
		slice := v.Convert(sSliceType).Interface().([]string)
		return reflect.ValueOf(utils.NewStringSet(slice...)), nil
	case v.Type().ConvertibleTo(iSliceType):
		slice := v.Convert(iSliceType).Interface().([]interface{})
		set := utils.NewStringSetFromSet(utils.NewSet(slice...))
		return reflect.ValueOf(set), nil
	default:
		return reflect.Value{}, fmt.Errorf("type %T cannot be converted to timestamp", v.Interface())
	}
}

func toSet(v reflect.Value) (reflect.Value, error) {
	switch {
	case v.Type().ConvertibleTo(sSliceType):
		slice := v.Convert(sSliceType).Interface().([]string)
		return reflect.ValueOf(utils.NewStringSet(slice...).ToSet()), nil
	case v.Type().ConvertibleTo(iSliceType):
		slice := v.Convert(iSliceType).Interface().([]interface{})
		return reflect.ValueOf(utils.NewSet(slice...)), nil
	default:
		return reflect.Value{}, fmt.Errorf("type %T cannot be converted to timestamp", v.Interface())
	}
}
