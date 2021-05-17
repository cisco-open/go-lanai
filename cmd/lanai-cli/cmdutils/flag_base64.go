package cmdutils

import "encoding/base64"

// Base64Value implements pflag.Value and pflag.SliceValue
type Base64Value struct {
	ptr *[]byte
}

func newBase64Value(defaultVal []byte, p *[]byte) *Base64Value {
	if p != nil {
		*p = defaultVal
	}
	return &Base64Value{
		ptr: p,
	}
}

// pflag.Value
func (v Base64Value) String() string {
	if v.ptr == nil {
		return "nil"
	}
	return base64.StdEncoding.EncodeToString(*v.ptr)
}

func (v *Base64Value) Set(s string) error {
	data, e := v.decode(s)
	if e != nil {
		return e
	}
	v.ptr = &data
	return nil
}

func (v Base64Value) Type() string {
	return "base64"
}

//// pflag.SliceValue
//func (v *Base64Value) Append(s string) error {
//	data, e := v.decode(s)
//	if e != nil {
//		return e
//	}
//	v.Data = append(v.Data, data...)
//	return nil
//}
//
//func (v *Base64Value) Replace(strings []string) error {
//	for _, s := range strings {
//		data, e := v.decode(s)
//		if e != nil {
//			return e
//		}
//		v.Data = append(v.Data, data...)
//	}
//	return nil
//}
//
//func (v *Base64Value) GetSlice() []string {
//	panic("implement me")
//}

func (v Base64Value) decode(s string) ([]byte, error) {
	var data []byte
	_, e := base64.StdEncoding.Decode([]byte(s), data)
	return data, e
}
