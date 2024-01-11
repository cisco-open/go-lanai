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

func (s StringSet) HasAll(values ...string) bool {
	for _, v := range values {
		if !s.Has(v) {
			return false
		}
	}
	return true
}

func (s StringSet) Equals(another StringSet) bool {
	if len(s) != len(another){
		return false
	} else if len(s) == 0 && len(another) == 0 {
		return true
	}
	for k := range another {
		if !s.Has(k) {
			return false
		}
	}
	return true
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
	cp := NewStringSet()
	for k,_ := range s {
		cp[k] = void{}
	}
	return cp
}

func (s StringSet) ToSet() Set {
	return NewSetFromStringSet(s)
}

// MarshalJSON json.Marshaler
func (s StringSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

// UnmarshalJSON json.Unmarshaler
func (s StringSet) UnmarshalJSON(data []byte) error {
	values := make([]string, 0)
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	if s != nil {
		s.Add(values...)
	}
	return nil
}

/** Interface Set **/

type Set map[interface{}]void

func NewSet(values...interface{}) Set {
	return make(Set).Add(values...)
}

func NewSetFromStringSet(stringSet StringSet) Set {
	set := NewSet()
	for k := range stringSet {
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
		slice := make([]interface{}, 0)
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

func (s Set) HasAll(values ...interface{}) bool {
	for _, v := range values {
		if !s.Has(v) {
			return false
		}
	}
	return true
}

func (s Set) Equals(another Set) bool {
	if len(s) != len(another){
		return false
	} else if len(s) == 0 && len(another) == 0 {
		return true
	}
	for k := range another {
		if !s.Has(k) {
			return false
		}
	}
	return true
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
	cp := NewSet()
	for k := range s {
		cp[k] = void{}
	}
	return cp
}

// MarshalJSON json.Marshaler
func (s Set) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

// UnmarshalJSON json.Unmarshaler
func (s Set) UnmarshalJSON(data []byte) error {
	values := make([]interface{}, 0)
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	if s != nil {
		s.Add(values...)
	}
	return nil
}

/** Generic Set **/

type GenericSet[T comparable] map[T]void

func NewGenericSet[T comparable](values...T) GenericSet[T] {
	return make(GenericSet[T]).Add(values...)
}

func (s GenericSet[T]) Add(values...T) GenericSet[T] {
	for _, item := range values {
		s[item] = void{}
	}
	return s
}

func (s GenericSet[T]) Remove(values...T) GenericSet[T] {
	for _, item := range values {
		delete(s, item)
	}
	return s
}

func (s GenericSet[T]) Has(value T) bool {
	_, ok := s[value]
	return ok
}

func (s GenericSet[T]) HasAll(values ...T) bool {
	for _, v := range values {
		if !s.Has(v) {
			return false
		}
	}
	return true
}

func (s GenericSet[T]) Equals(another GenericSet[T]) bool {
	if len(s) != len(another){
		return false
	} else if len(s) == 0 && len(another) == 0 {
		return true
	}
	for k := range another {
		if !s.Has(k) {
			return false
		}
	}
	return true
}

func (s GenericSet[T]) Values() []T {
	values := make([]T, len(s))
	var i int
	for item := range s {
		values[i] = item
		i++
	}
	return values
}

func (s GenericSet[T]) Copy() GenericSet[T] {
	cp := NewGenericSet[T]()
	for k := range s {
		cp[k] = void{}
	}
	return cp
}

// MarshalJSON json.Marshaler
func (s GenericSet[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

// UnmarshalJSON json.Unmarshaler
func (s GenericSet[T]) UnmarshalJSON(data []byte) error {
	values := make([]T, 0)
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	if s != nil {
		s.Add(values...)
	}
	return nil
}