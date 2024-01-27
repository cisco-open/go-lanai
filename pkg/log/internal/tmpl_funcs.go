// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

// Note: https://pkg.go.dev/text/template#hdr-Pipelines chainable argument should be the last parameter of any function
var (
	TmplFuncMap = template.FuncMap{
		"cap":   Capped,
		"pad":   Padding,
		"lvl":   MakeLevelFunc(true),
		"join":  Join,
		"trace": Trace,
	}
	TmplFuncMapNonTerm = template.FuncMap{
		"cap":   Capped,
		"pad":   Padding,
		"lvl":   MakeLevelFunc(false),
		"join":  Join,
		"trace": Trace,
	}
)

type levelFuncs struct {
	text  func(int) string
	color func(interface{}) string
}

var (
	levelFuncsMap = map[string]levelFuncs{
		"debug": {
			text: MakeLevelPaddingFunc("DEBUG"),
			color: MakeQuickColorFunc(Gray),
		},
		"info": {
			text: MakeLevelPaddingFunc("INFO"),
			color: MakeQuickColorFunc(Cyan),
		},
		"warn": {
			text: MakeLevelPaddingFunc("WARN"),
			color: MakeQuickColorFunc(BoldYellow),
		},
		"error": {
			text: MakeLevelPaddingFunc("ERROR"),
			color: MakeQuickColorFunc(BoldRed),
		},
	}
)

func MakeKVFunc(ignored utils.StringSet) func(Fields) string {
	return func(kvs Fields) string {
		kvStrs := make([]string, 0, len(kvs))
		for k, v := range kvs {
			if v == nil || ignored.Has(k) || reflect.ValueOf(v).IsZero() {
				continue
			}
			kvStrs = append(kvStrs, fmt.Sprintf(`%s="%v"`, k, v))
		}

		if len(kvStrs) == 0 {
			return ""
		}
		return "{" + strings.Join(kvStrs, ", ") + "}"
	}
}

func MakeLevelFunc(term bool) func(padding int, kvs Fields) string {
	if term {
		return func(padding int, kvs Fields) string {
			lv, _ := kvs[LogKeyLevel]
			lvStr := Sprint(lv)
			if funcs, ok := levelFuncsMap[lvStr]; ok {
				return funcs.color(funcs.text(padding))
			}
			return lvStr
		}
	} else {
		return func(padding int, kvs Fields) string {
			lv, _ := kvs[LogKeyLevel]
			lvStr := Sprint(lv)
			if funcs, ok := levelFuncsMap[lvStr]; ok {
				return funcs.text(padding)
			}
			return lvStr
		}
	}
}

func MakeLevelPaddingFunc(v interface{}) func(int) string {
	return func(p int) string {
		return Padding(p, v)
	}
}

// Padding example: `{{padding -6 value}}` "{{padding 10 value}}"
func Padding(padding int, v interface{}) string {
	tag := "%" + strconv.Itoa(padding) + "v"
	return fmt.Sprintf(tag, v)
}

// Capped truncate given value to specified length
// if cap > 0: with tailing "..." if truncated
// if cap < 0: with middle "..." if truncated
func Capped(cap int, v interface{}) string {
	c := int(math.Abs(float64(cap)))
	s := Sprint(v)
	if len(s) <= c {
		return s
	}
	if cap > 0 {
		return fmt.Sprintf("%." + strconv.Itoa(c - 3) + "s...", s)
	} else if cap < 0 {
		lead := (c - 3) / 2
		tail := c - lead - 3
		return fmt.Sprintf("%." + strconv.Itoa(lead) + "s...%s", s, s[len(s)-tail:])
	} else {
		return ""
	}
}

func Join(sep string, values ...interface{}) string {
	strs := make([]string, 0, len(values))
	for _, v := range values {
		s := Sprint(v)
		if s != "" {
			strs = append(strs, s)
		}
	}
	str := strings.Join(strs, sep)
	return str
}

// Trace generate shortest possible tracing info string:
// 	- if trace ID is not available, return empty string
//  - if span ID is same as trace ID, we assume parent ID is 0 and only returns traceID
//  - if span ID is different from trace ID and parent ID is same as trace ID, we only returns trace ID and span ID
func Trace(tid, sid, pid interface{}) string {
	tidStr, sidStr, pidStr := Sprint(tid), Sprint(sid), Sprint(pid)
	switch {
	case tidStr == "":
		return ""
	case sidStr == tidStr:
		return tidStr
	case pidStr == tidStr:
		return tidStr + " " + sidStr
	default:
		return tidStr + " " + sidStr + " " + pidStr
	}
}

func Sprint(v interface{}) string {
	switch v.(type) {
	case nil:
		return ""
	case string:
		return v.(string)
	case []byte:
		return string(v.([]byte))
	case fmt.Stringer:
		return v.(fmt.Stringer).String()
	case encoding.TextMarshaler:
		if s, e := v.(encoding.TextMarshaler).MarshalText(); e == nil {
			return string(s)
		}
	}
	return fmt.Sprintf("%v", v)
}
