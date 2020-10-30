package appconfig

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
)

//TODO: when unflatten, we don't support the case where the key contains an index yet.

// Options the flatten options.
// By default: Demiliter = "."
type Options struct {
	Delimiter string
}

func VisitEach(nested map[string]interface{}, apply func(string, interface{}), configures...func(*Options)) error {
	opts := &Options{
		Delimiter: ".",
	}

	for _, configure := range configures {
		configure(opts)
	}

	return recursiveVisit("", nested, apply, opts)
}

func recursiveVisit(key string, value interface{}, apply func(string, interface{}), opts *Options) (err error) {
	switch value := value.(type) {
	case map[string]interface{}:
		//if empty map, can't do anything
		if reflect.DeepEqual(value, map[string]interface{}{}) {
			return
		}
		for k, v := range value {
			// create new key
			newKey := k
			if key != "" {
				newKey = key + opts.Delimiter + newKey
			}
			fe := recursiveVisit(newKey, v, apply, opts)
			if fe != nil {
				err = fe
				return
			}
		}
	case []interface{}:
		//if empty slice
		if reflect.DeepEqual(value, []interface{}{}) {
			return
		}
		for i, v := range value {
			newKey := strconv.Itoa(i)
			if key != "" {
				newKey = key + opts.Delimiter + newKey
			}
			fe := recursiveVisit(newKey, v, apply, opts)
			if fe != nil {
				err = fe
				return
			}
		}
	default:
		apply(key, value)
	}
	return
}

//TODO: snake case

// Unflatten the map, it returns a nested map of a map
// By default, the flatten has Delimiter = "."
func UnFlatten(flat map[string]interface{}, configures...func(*Options)) (nested map[string]interface{}, err error) {
	opts := &Options{
		Delimiter: ".",
	}

	for _, configure := range configures {
		configure(opts)
	}

	nested = make(map[string]interface{})

	for k, v := range flat {
		temp := uf(k, v, opts).(map[string]interface{})
		err = mergo.Merge(&nested, temp)
		if err != nil {
			return
		}
	}

	return
}


func uf(k string, v interface{}, opts *Options) (n interface{}) {
	n = v

	keys := strings.Split(k, opts.Delimiter)

	for i := len(keys) - 1; i >= 0; i-- {
		temp := make(map[string]interface{})
		temp[keys[i]] = n
		n = temp
	}

	return
}

func UnFlattenKey(k string, configures...func(*Options)) []string {
	//TODO: deduplicate
	opts := &Options{
		Delimiter: ".",
	}

	for _, configure := range configures {
		configure(opts)
	}

	return strings.Split(k, opts.Delimiter)
}