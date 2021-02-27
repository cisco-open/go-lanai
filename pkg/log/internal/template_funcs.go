package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding"
	"fmt"
	"github.com/go-kit/kit/log/level"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"text/template"
)

var TmplFuncMap = template.FuncMap{
	"padding": Padding,
	"level":   Level,
	"red":     Red,
	"green":   Green,
	"gray":    Gray,
	"cyan":    Cyan,
	"yellow":  Yellow,
	"white":   White,
	"blue":    Blue,
}

type levelFunc func(int) string

var (
	levelFuncs = map[string]levelFunc{
		"debug": func(p int) string {
			return Gray(Padding("DEBUG", p))
		},
		"info": func(p int) string {
			return Cyan(Padding("INFO", p))
		},
		"warn": func(p int) string {
			return Yellow(Padding("WARN", p))
		},
		"error": func(p int) string {
			return Red(Padding("ERROR", p))
		},
	}
)

var (
	noop   = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	gray   = "\033[37m"
	white  = "\033[97m"
)

func init() {
	if runtime.GOOS == "windows" {
		noop = ""
		red = ""
		green = ""
		yellow = ""
		blue = ""
		purple = ""
		cyan = ""
		gray = ""
		white = ""
	}
}

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

// padding example: `{{padding value -6}}` "{{padding value 10}}"
func Padding(v interface{}, padding int) string {
	tag := "%" + strconv.Itoa(padding) + "v"
	return fmt.Sprintf(tag, v)
}

func Level(kvs Fields, padding int) string {
	lv, ok := kvs[level.Key().(string)]
	if !ok {
		return ""
	}

	f, ok := levelFuncs[Sprint(lv)]
	if !ok {
		return ""
	}
	return f(padding)
}

func Red(s interface{}) string {

	return red + Sprint(s) + noop
}

func Green(s interface{}) string {
	return green + Sprint(s) + noop
}

func Gray(s interface{}) string {
	return gray + Sprint(s) + noop
}

func Cyan(s interface{}) string {
	return cyan + Sprint(s) + noop
}

func Yellow(s interface{}) string {
	return yellow + Sprint(s) + noop
}

func White(s interface{}) string {
	return white + Sprint(s) + noop
}

func Blue(s interface{}) string {
	return blue + Sprint(s) + noop
}

func Sprint(v interface{}) string {
	switch v.(type) {
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
