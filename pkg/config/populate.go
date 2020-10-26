//TODO: revisit the implementation here, compare it to the implementation in golang json library that populates struct
// from json data.

// Copyright 2018 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PartialConfig struct {
	local  map[string]string
	config *Config
}

func (p PartialConfig) Get(key string) (string, bool) {
	normalizedKey := NormalizeKey(key)
	value, ok := p.local[normalizedKey]
	return value, ok
}

func (p PartialConfig) Set(key, value string) {
	normalizedKey := NormalizeKey(key)
	p.local[normalizedKey] = value
}

// FilterStripPrefix returns a new properties object with a subset of all keys
// with the given prefix and the prefix removed from the keys.
func (p PartialConfig) FilterStripPrefix(prefix string) PartialConfig {
	pp := PartialConfig{
		local:  make(map[string]string),
		config: p.config,
	}
	n := len(prefix)
	for k := range p.local {
		if len(k) > len(prefix) && strings.HasPrefix(k, prefix) {
			// TODO(fs): we are ignoring the error which flags a circular reference.
			// TODO(fs): since we are modifying keys I am not entirely sure whether we can create a circular reference
			// TODO(fs): this function should probably return an error but the signature is fixed
			newk := k[n:]
			if strings.HasPrefix(newk, ".") {
				newk = newk[1:]
			}
			pp.Set(newk, p.local[k])
		}
	}
	return pp
}

func (p PartialConfig) HasKey(prefix string) bool {
	for k := range p.local {
		if k == prefix {
			return true
		} else if strings.HasPrefix(k, prefix+".") {
			return true
		} else if strings.HasPrefix(k, prefix+"[") {
			return true
		}
	}
	return false
}

// Populate assigns property values to exported fields of a struct.
//
// Populate traverses v recursively and returns an error if a value cannot be
// converted to the field type or a required value is missing for a field.
//
// The following type dependent decodings are used:
//
// String, boolean, numeric fields have the value of the property key assigned.
// The property key name is the name of the field. A different key and a default
// value can be set in the field's tag. Fields without default value are
// required. If the value cannot be converted to the field type an error is
// returned.
//
// time.Duration fields have the result of time.ParseDuration() assigned.
//
// time.Time fields have the vaule of time.Parse() assigned. The default layout
// is time.RFC3339 but can be set in the field's tag.
//
// Arrays and slices of string, boolean, numeric, time.Duration and time.Time
// fields have the value interpreted as a comma separated list of values. The
// individual values are trimmed of whitespace and empty values are ignored. A
// default value can be provided as a semicolon separated list in the field's
// tag.
//
// Struct fields are decoded recursively using the field name plus "." as
// prefix. The prefix (without dot) can be overridden in the field's tag.
// Default values are not supported in the field's tag. Specify them on the
// fields of the inner struct instead.
//
// Map fields must have a key of type string and are decoded recursively by
// using the field's name plus ".' as prefix and the next element of the key
// name as map key. The prefix (without dot) can be overridden in the field's
// tag. Default values are not supported.
//
// Examples:
//
//     // Field is ignored.
//     Field int `config:"-"`
//
//     // Field is assigned value of 'Field'.
//     Field int
//
//     // Field is assigned value of 'myName'.
//     Field int `config:"myName"`
//
//     // Field is assigned value of key 'myName' and has a default
//     // value 15 if the key does not exist.
//     Field int `config:"myName,default=15"`
//
//     // Field is assigned value of key 'Field' and has a default
//     // value 15 if the key does not exist.
//     Field int `config:"default=15"`
//
//     // Field is assigned value of key 'date' and the date
//     // is in format 2006-01-02
//     Field time.Time `config:"date,layout=2006-01-02"`
//
//     // Field is assigned the non-empty and whitespace trimmed
//     // values of key 'Field' split by commas.
//     Field []string
//
//     // Field is assigned the non-empty and whitespace trimmed
//     // values of key 'Field' split by commas and has a default
//     // value ["a", "b", "c"] if the key does not exist.
//     Field []string `config:"default=a;b;c"`
//
//     // Field is decoded recursively with "Field." as key prefix.
//     Field SomeStruct
//
//     // Field is decoded recursively with "myName." as key prefix.
//     Field SomeStruct `config:"myName"`
//
//     // Field is decoded recursively with "Field." as key prefix
//     // and the next dotted element of the key as map key.
//     Field map[string]string
//
//     // Field is decoded recursively with "myName." as key prefix
//     // and the next dotted element of the key as map key.
//     Field map[string]string `config:"myName"`
func (p PartialConfig) Populate(x interface{}) error {
	t, v := reflect.TypeOf(x), reflect.ValueOf(x)
	if t.Kind() != reflect.Ptr {
		return errors.Errorf("Cannot populate config: not a pointer: %q", t)
	}
	if err := dec(p, "", nil, nil, v); err != nil {
		return err
	}
	return nil
}

func NewPartialConfig(local map[string]string, config *Config) PartialConfig {
	return PartialConfig{
		local:  local,
		config: config,
	}
}

func dec(p PartialConfig, key string, def *string, opts map[string]string, v reflect.Value) error {
	t := v.Type()
	key = NormalizeKey(key)

	// value returns the property value for key or the default if provided.
	value := func() (string, error) {
		if val, ok := p.Get(key); ok {
			return val, nil
		}
		if def != nil {
			return p.config.resolveValue(make(map[string]interface{}), p.config.settings, *def), nil
		}
		return "", fmt.Errorf("missing required key %s", key)
	}

	arrayStructValue := func() ([]PartialConfig, error) {
		values := []*PartialConfig{}

		keyPattern := strings.ReplaceAll(key, ".", "\\.") + `\[(\d+)\]`
		keyRegex, _ := regexp.Compile(keyPattern)
		for k, v := range p.local {
			// Parse the key for a "key[index]" match
			matches := keyRegex.FindStringSubmatch(k)
			if len(matches) != 2 {
				continue
			}

			// Extract the index
			index, err := strconv.Atoi(matches[1])
			if err != nil {
				fmt.Println("Invalid property index: %s", k)
				continue
			}

			// Ensure we have enough slots in the slice
			for index >= len(values) {
				values = append(values, nil)
			}

			// Set the value into the slice
			submap := values[index]
			if submap == nil {
				submap = &PartialConfig{
					local:  make(map[string]string),
					config: p.config,
				}
				values[index] = submap
			}

			submapKey := strings.TrimPrefix(k, fmt.Sprint(key, "[", index, "]."))
			submap.local[submapKey] = v
		}

		var result []PartialConfig
		for _, v := range values {
			if v != nil {
				result = append(result, *v)
			}
		}

		return result, nil
	}

	arrayValue := func() ([]string, error) {
		values := []*string{}

		// Check the single setting first
		if flattenedValue, ok := p.Get(key); ok {
			// Comma-separated values from spring
			splitValues := split(flattenedValue, ",")
			for _, v := range splitValues {
				value := v
				values = append(values, &value)
			}
		}

		keyPattern := strings.ReplaceAll(key, ".", "\\.") + `\[(\d+)\]`
		keyRegex, _ := regexp.Compile(keyPattern)
		for k, v := range p.local {
			// Parse the key for a "key[index]" match
			matches := keyRegex.FindStringSubmatch(k)
			if len(matches) != 2 {
				continue
			}

			// Extract the index
			index, err := strconv.Atoi(matches[1])
			if err != nil {
				fmt.Println("Invalid property index: %s", k)
				continue
			}

			// Ensure we have enough slots in the slice
			for index >= len(values) {
				values = append(values, nil)
			}

			// Set the value into the slice
			v2 := v
			values[index] = &v2
		}

		var result []string
		for _, v := range values {
			if v != nil {
				result = append(result, *v)
			}
		}

		if len(result) == 0 {
			if def != nil {
				defResolved := p.config.resolveValue(make(map[string]interface{}), p.config.settings, *def)
				return strings.Split(defResolved, ";"), nil
			}
		}

		return result, nil
	}

	// conv converts a string to a value of the given type.
	conv := func(s string, t reflect.Type) (val reflect.Value, err error) {
		var v interface{}

		switch {
		case isDuration(t):
			v, err = time.ParseDuration(s)

		case isTime(t):
			layout := opts["layout"]
			if layout == "" {
				layout = time.RFC3339
			}
			v, err = time.Parse(layout, s)

		case isBool(t):
			v, err = boolVal(s), nil

		case isString(t):
			v, err = s, nil

		case isFloat(t):
			v, err = strconv.ParseFloat(s, 64)

		case isInt(t):
			v, err = strconv.ParseInt(s, 10, 64)

		case isUint(t):
			v, err = strconv.ParseUint(s, 10, 64)

		default:
			return reflect.Zero(t), fmt.Errorf("unsupported type %s", t)
		}
		if err != nil {
			return reflect.Zero(t), err
		}
		return reflect.ValueOf(v).Convert(t), nil
	}

	// keydef returns the property key and the default value based on the
	// name of the struct field and the options in the tag.
	keydef := func(f reflect.StructField) (string, *string, map[string]string) {
		_key, _opts := parseTag(f.Tag.Get("config"))

		var _def *string
		if d, ok := _opts["default"]; ok {
			_def = &d
		}
		if _key != "" {
			return _key, _def, _opts
		}
		return f.Name, _def, _opts
	}

	switch {
	case isDuration(t) || isTime(t) || isBool(t) || isString(t) || isFloat(t) || isInt(t) || isUint(t):
		s, err := value()
		if err != nil {
			return err
		}
		val, err := conv(s, t)
		if err != nil {
			return err
		}
		v.Set(val)

	case isPtr(t):
		if v.IsNil() {
			if p.HasKey(key) {
				elem := reflect.New(t.Elem())
				v.Set(elem)
			} else {
				return nil
			}
		}
		return dec(p, key, def, opts, v.Elem())

	case isStruct(t):
		for i := 0; i < v.NumField(); i++ {
			fv := v.Field(i)
			fk, def, opts := keydef(t.Field(i))
			if !fv.CanSet() {
				return fmt.Errorf("cannot set %s", t.Field(i).Name)
			}
			if fk == "-" {
				continue
			}
			if key != "" {
				fk = key + "." + fk
			}
			if err := dec(p, fk, def, opts, fv); err != nil {
				return err
			}
		}
		return nil

	case isArray(t):
		if isStruct(t.Elem()) {
			partialConfigs, err := arrayStructValue()
			if err != nil {
				return err
			}

			a := reflect.MakeSlice(t, 0, len(partialConfigs))
			for _, partialConfig := range partialConfigs {
				v := reflect.New(t.Elem())
				if err := dec(partialConfig, "", nil, nil, v); err != nil {
					return err
				}
				a = reflect.Append(a, v.Elem())
			}
			v.Set(a)

		} else {
			// Scalars
			vals, err := arrayValue()
			if err != nil {
				return err
			}
			a := reflect.MakeSlice(t, 0, len(vals))
			for _, s := range vals {
				val, err := conv(s, t.Elem())
				if err != nil {
					return err
				}
				a = reflect.Append(a, val)
			}
			v.Set(a)
		}

	case isMap(t):
		valT := t.Elem()
		m := reflect.MakeMap(t)
		for postfix := range p.FilterStripPrefix(key + ".").local {
			pp := strings.SplitN(postfix, ".", 2)
			mk, mv := pp[0], reflect.New(valT)
			if err := dec(p, key+"."+mk, nil, nil, mv); err != nil {
				return err
			}
			m.SetMapIndex(reflect.ValueOf(mk), mv.Elem())
		}
		v.Set(m)

	default:
		return fmt.Errorf("unsupported type %s", t)
	}
	return nil
}

// split splits a string on sep, trims whitespace of elements
// and omits empty elements
func split(s string, sep string) []string {
	var a []string
	for _, v := range strings.Split(s, sep) {
		if v = strings.TrimSpace(v); v != "" {
			a = append(a, v)
		}
	}
	return a
}

// parseTag parses a "key,k=v,k=v,..."
func parseTag(tag string) (key string, opts map[string]string) {
	opts = map[string]string{}
	for i, s := range strings.Split(tag, ",") {
		pp := strings.SplitN(s, "=", 2)
		if len(pp) == 1 {
			if i == 0 {
				key = s
			} else {
				opts[pp[0]] = ""
			}
		} else {
			opts[pp[0]] = pp[1]
		}
	}
	return key, opts
}

func isArray(t reflect.Type) bool    { return t.Kind() == reflect.Array || t.Kind() == reflect.Slice }
func isBool(t reflect.Type) bool     { return t.Kind() == reflect.Bool }
func isDuration(t reflect.Type) bool { return t == reflect.TypeOf(time.Second) }
func isMap(t reflect.Type) bool      { return t.Kind() == reflect.Map }
func isPtr(t reflect.Type) bool      { return t.Kind() == reflect.Ptr }
func isString(t reflect.Type) bool   { return t.Kind() == reflect.String }
func isStruct(t reflect.Type) bool   { return t.Kind() == reflect.Struct }
func isTime(t reflect.Type) bool     { return t == reflect.TypeOf(time.Time{}) }
func isFloat(t reflect.Type) bool {
	return t.Kind() == reflect.Float32 || t.Kind() == reflect.Float64
}
func isInt(t reflect.Type) bool {
	return t.Kind() == reflect.Int || t.Kind() == reflect.Int8 || t.Kind() == reflect.Int16 || t.Kind() == reflect.Int32 || t.Kind() == reflect.Int64
}
func isUint(t reflect.Type) bool {
	return t.Kind() == reflect.Uint || t.Kind() == reflect.Uint8 || t.Kind() == reflect.Uint16 || t.Kind() == reflect.Uint32 || t.Kind() == reflect.Uint64
}

func boolVal(v string) bool {
	v = strings.ToLower(v)
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
