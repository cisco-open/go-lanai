package utils

import (
	"encoding/json"
	"fmt"
)

type void struct{}

/** StringSet **/
type StringSet map[string]void

func NewStringSet(values...string) StringSet {
	return make(StringSet).Add(values...)
}

func NewStringSetFromSet(set Set) StringSet {
	stringSet := make(StringSet)
	for k, _ := range set {
		if str, ok := k.(string); ok {
			stringSet[str] = void{}
		}
	}
	return stringSet
}

func NewStringSetFrom(i interface{}) StringSet {
	switch i.(type) {
	case StringSet:
		return i.(StringSet).Copy()
	case Set:
		return NewStringSetFromSet(i.(Set))
	case []string:
		return NewStringSet(i.([]string)...)
	case []interface{}:
		slice := []string{}
		for _,v := range i.([]interface{}) {
			if s,ok := v.(string); ok {
				slice = append(slice, s)
			}
		}
		return NewStringSet(slice...)
	default:
		panic(fmt.Errorf("new StringSet from unsupported type %T", i))
	}
}

func (s StringSet) Add(values...string) StringSet {
	for _, item := range values {
		s[item] = void{}
	}
	return s
}

func (s StringSet) Remove(values...string) StringSet {
	for _, item := range values {
		delete(s, item)
	}
	return s
}

func (s StringSet) Has(value string) bool {
	_, ok := s[value]
	return ok
}

func (s StringSet) Values() []string {
	values := make([]string, len(s))
	var i int
	for item := range s {
		values[i] = item
		i++
	}
	return values
}

func (s StringSet) Copy() StringSet {
	copy := NewStringSet()
	for k,_ := range s {
		copy[k] = void{}
	}
	return copy
}

func (s StringSet) ToSet() Set {
	return NewSetFromStringSet(s)
}

// json.Marshaler
func (s StringSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

// json.Unmarshaler
func (s StringSet) UnmarshalJSON(data []byte) error {
	values := []string{}
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	s.Add(values...)
	return nil
}

/** Generic Set **/
type Set map[interface{}]void

func NewSet(values...interface{}) Set {
	return make(Set).Add(values...)
}

func NewSetFromStringSet(stringSet StringSet) Set {
	set := NewSet()
	for k, _ := range stringSet {
		set[k] = void{}
	}
	return set
}

func NewSetFrom(i interface{}) Set {
	switch i.(type) {
	case StringSet:
		return NewSetFromStringSet(i.(StringSet))
	case Set:
		return i.(Set).Copy()
	case []string:
		slice := []interface{}{}
		for _,v := range i.([]string) {
			slice = append(slice, v)
		}
		return NewSet(slice...)
	case []interface{}:
		return NewSet(i.([]interface{})...)
	default:
		panic(fmt.Errorf("new StringSet from unsupported type %T", i))
	}
}

func (s Set) Add(values...interface{}) Set {
	for _, item := range values {
		s[item] = void{}
	}
	return s
}

func (s Set) Remove(values...interface{}) Set {
	for _, item := range values {
		delete(s, item)
	}
	return s
}

func (s Set) Has(value interface{}) bool {
	_, ok := s[value]
	return ok
}

func (s Set) Values() []interface{} {
	values := make([]interface{}, len(s))
	var i int
	for item := range s {
		values[i] = item
		i++
	}
	return values
}

func (s Set) Copy() Set {
	copy := NewSet()
	for k,_ := range s {
		copy[k] = void{}
	}
	return copy
}

// json.Marshaler
func (s Set) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

// json.Unmarshaler
func (s Set) UnmarshalJSON(data []byte) error {
	values := []interface{}{}
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	s.Add(values...)
	return nil
}