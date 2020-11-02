package appconfig

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/imdario/mergo"
)

// Options the flatten options.
// By default: Demiliter = "."
type Options struct {
	Delimiter string
}

func ProcessKeyFormat(nested map[string]interface{}, processor func(string, ...func(*Options)) string) (map[string]interface{}, error) {
	result, err := processKeyFormat(nested, processor)

	return result.(map[string]interface{}), err
}

func processKeyFormat(value interface{}, processor func(string, ...func(*Options)) string) (interface{}, error) {

	switch value := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		//if empty map, can't do anything
		if reflect.DeepEqual(value, map[string]interface{}{}) {
			return result, nil
		}
		for k, v := range value {
			newKey := processor(k)
			visitedValue, fe := processKeyFormat(v, processor)
			if fe != nil {
				return nil, fe
			}
			result[newKey] = visitedValue
		}
		return result, nil
	case []interface{}:
		//if empty slice
		var result []interface{}
		if reflect.DeepEqual(value, []interface{}{}) {
			return result, nil
		}
		for _, v := range value {
			visitedValue, fe := processKeyFormat(v, processor)
			if fe != nil {
				return nil, fe
			}
			result = append(result, visitedValue)
		}
		return result, nil
	default:
		return value, nil
	}
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

//We don't support key with index number at the end here.
//TODO: Because it is only useful for system properties to override list. So it should be taken care of by the environment properties provider
//  ie:
//  spring.my-example.url[0]=https://example.com
// or
//  spring.my-example.url=https://example.com,https://spring.io
// Therefore we need to:
//   1. check if key ends with []
//   2. hash them into the comma format
//   3. process the commas separately

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
	opts := &Options{
		Delimiter: ".",
	}

	for _, configure := range configures {
		configure(opts)
	}

	return strings.Split(k, opts.Delimiter)
}

/*
Allow relaxed binding. The following should all bind to the same

acme.my-project.person.first-name
acme.myProject.person.firstName

we want to standardize on the snake case convention. So our algorithm is to
find the upper case letter, and if it's not preceded by a dash, then insert the dash

TODO: environment variables have ACME_MYPROJECT_PERSON_FIRSTNAME.
 This should be taken care of by the environment provider first to acme.myproject.person.firstname

*/

const dash = rune('-')

func NormalizeKey(key string, configures...func(*Options)) string {
	keys := UnFlattenKey(key, configures...)

	result := ""

	for i, key := range keys {
		var normalized []rune
		for pos, char := range key {
			if unicode.IsUpper(char) {
				if pos>0 {
					normalized = append(normalized, dash)
				}
				normalized = append(normalized, unicode.ToLower(char))
			} else {
				normalized = append(normalized, char)
			}
		}

		result = result + string(normalized)
		if i < len(keys) - 1 {
			result = result + "."
		}
	}

	return result
}