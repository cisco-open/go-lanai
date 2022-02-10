package cmdutils

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"reflect"
	"strings"
)

const (
	TagKeyFlag = "flag"
	TagKeyDescription = "desc"
	TagFlagSeparator = ","
	TagFlagRequired = "required"
	//TagKeyEnv = "env"
)

type flagOptions func(*flagMeta) error

type flagMeta struct {
	fullName    string
	shortName   string
	description string
	required    bool
	defaultVal  interface{}
	ptr         interface{}
}

// PersistentFlags takes Struct or *Struct value and register its fields using "flag" tag with cmd.PersistentFlags
// it panic if given value is not a struct
// Tag syntax:
//		`flag:"<name>[,<short_name>][,required]" desc:"<description for help cmd>"`
func PersistentFlags(cmd *cobra.Command, value interface{}) {
	flags := parseForFlags(value)
	for _, meta := range flags {
		if e := registerFlags(cmd.PersistentFlags(), meta, requirePersistentFlag(cmd)); e != nil {
			panic(e)
		}
	}
}

// LocalFlags is similar to PersistentFlags
func LocalFlags(cmd *cobra.Command, value interface{}) {
	flags := parseForFlags(value)
	for _, meta := range flags {
		if e := registerFlags(cmd.PersistentFlags(), meta, requireLocalFlag(cmd)); e != nil {
			panic(e)
		}
	}
}

func parseForFlags(value interface{}) []*flagMeta {
	metas, e := parseStructForFlags(reflect.ValueOf(value))
	if e != nil {
		panic(e)
	}
	return metas
}

func parseStructForFlags(v reflect.Value) (ret []*flagMeta, err error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ret, fmt.Errorf("unsupported value %T", v.Interface())
	}

	ret = []*flagMeta{}
	t := v.Type()
	total := v.NumField()
	for i := 0; i < total; i++ {
		f := t.Field(i)
		fv := v.Field(i)

		if isStructOrPtr(f.Type) {
			metas, e := parseStructForFlags(fv)
			if e != nil {
				return nil, e
			}
			ret = append(ret, metas...)
			continue
		}

		meta, e := parseStructFieldForFlags(f, fv)
		if e != nil || meta == nil {
			return nil, e
		}
		ret = append(ret, meta)
	}
	return
}

func parseStructFieldForFlags(f reflect.StructField, fv reflect.Value) (*flagMeta, error) {
	flatTag, ok := f.Tag.Lookup(TagKeyFlag)
	if !ok {
		return nil, nil
	}

	dv, addr, e := resolveFieldValue(fv)
	if e != nil {
		return nil, fmt.Errorf(`invalid flag declaration at field [%s]: %v`, f.Name, e)
	}

	// get and process tag
	tokens := strings.Split(flatTag, TagFlagSeparator)
	tokens[0] = strings.ToLower(strings.TrimSpace(tokens[0]))

	// create flag meta
	meta := flagMeta{
		defaultVal: dv,
		ptr: addr.Interface(),
	}
	if tokens[0] != "" {
		meta.fullName = tokens[0]
	}

	if len(tokens) > 1 {
		meta.shortName = tokens[1]
	}

	if len(tokens) > 2 && tokens[2] == TagFlagRequired {
		meta.required = true
	}

	// description
	meta.description = f.Tag.Get(TagKeyDescription)

	// validate and return
	if meta.fullName == "" {
		return nil, fmt.Errorf(`invalid flag tag "%s", missing name`, flatTag)
	}
	return &meta, nil
}

// resolveFieldValue resolve actual non-ptr default value of the field and its ptr
func resolveFieldValue(fv reflect.Value) (defaultVal interface{}, addr reflect.Value, err error) {
	switch {
	case fv.Kind() == reflect.Ptr && fv.IsNil():
		return defaultVal, addr, fmt.Errorf("pointer field with nil value is not supported as a flag")
	case fv.Kind() == reflect.Ptr:
		addr = fv
	case fv.CanAddr():
		addr = fv.Addr()
	default:
		return defaultVal, addr, fmt.Errorf("cannot get address")
	}

	defaultVal = addr.Elem().Interface()
	return
}

func isStructOrPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Struct || t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func registerFlags(flags *pflag.FlagSet, meta *flagMeta, options...flagOptions) error {
	switch p := meta.ptr.(type) {
	case *string:
		flags.StringVarP(p, meta.fullName, meta.shortName, meta.defaultVal.(string), meta.description)
	case *[]string:
		flags.StringSliceVarP(p, meta.fullName, meta.shortName, meta.defaultVal.([]string), meta.description)
	case *int:
		flags.IntVarP(p, meta.fullName, meta.shortName, meta.defaultVal.(int), meta.description)
	case *[]int:
		flags.IntSliceVarP(p, meta.fullName, meta.shortName, meta.defaultVal.([]int), meta.description)
	case *bool:
		flags.BoolVarP(p, meta.fullName, meta.shortName, meta.defaultVal.(bool), meta.description)
	case *[]bool:
		flags.BoolSliceVarP(p, meta.fullName, meta.shortName, meta.defaultVal.([]bool), meta.description)
	case *[]byte:
		flags.VarP(newBase64Value(meta.defaultVal.([]byte), p), meta.fullName, meta.shortName, meta.description)
	default:
		return fmt.Errorf(`unsupported type [%T] as flag`, meta.ptr)
	}

	for _, f := range options {
		if f == nil {
			continue
		}
		if e := f(meta); e != nil {
			return e
		}
	}
	return nil
}

func requirePersistentFlag(cmd *cobra.Command) flagOptions {
	return func(meta *flagMeta) error {
		if !meta.required {
			return nil
		}
		return cmd.MarkPersistentFlagRequired(meta.fullName)
	}
}

func requireLocalFlag(cmd *cobra.Command) flagOptions {
	return func(meta *flagMeta) error {
		if !meta.required {
			return nil
		}
		return cmd.MarkFlagRequired(meta.fullName)
	}
}
