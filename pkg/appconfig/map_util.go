package appconfig

import (
	"github.com/pkg/errors"
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

func VisitEach(nested map[string]interface{}, apply func(string, interface{}) error, configures...func(*Options)) error {
	opts := &Options{
		Delimiter: ".",
	}

	for _, configure := range configures {
		configure(opts)
	}

	return recursiveVisit("", nested, apply, opts)
}

//the recursive visit stops at the first error
func recursiveVisit(key string, value interface{}, apply func(string, interface{}) error, opts *Options) (err error) {
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
			newKey := "[" + strconv.Itoa(i) + "]"
			if key != "" {
				newKey = key + newKey
			}
			fe := recursiveVisit(newKey, v, apply, opts)
			if fe != nil {
				err = fe
				return
			}
		}
	default:
		err = apply(key, value)
	}
	return
}

type UfOptions struct {
	Delimiter   string
	AppendSlice bool
}

// This method supports un-flattening keys with index like the following
//  my-example.url[0]=https://example.com
// The indexed entries are treated like an unsorted list. The result will be a list but the order is not
// guaranteed to reflect the index order.
// A key with multiple index (a.b[0].c[0) is not supported
func UnFlatten(flat map[string]interface{}, configures...func(*UfOptions)) (nested map[string]interface{}, err error) {
	opts := &UfOptions{
		Delimiter:   ".",
		AppendSlice: true,
	}

	for _, configure := range configures {
		configure(opts)
	}

	nested = make(map[string]interface{})

	for k, v := range flat {
		temp, error := uf(k, v, opts)
		if error != nil {
			return nil, errors.Wrap(error, "cannot un-flatten due to error in key: " + k)
		}
		err = mergo.Merge(&nested, temp, func(c *mergo.Config) {
			c.AppendSlice = opts.AppendSlice
		})
		if err != nil {
			return
		}
	}

	return
}


func uf(k string, v interface{}, opts *UfOptions) (n interface{}, err error) {
	indexOccurance := 0

	n = v

	keys := strings.Split(k, opts.Delimiter)

	for i := len(keys) - 1; i >= 0; i-- {
		currKey := keys[i]
		temp := make(map[string]interface{})

		bracketLeft := strings.Index(currKey, "[")
		bracketRight := strings.Index(currKey, "]")

		if bracketLeft > 0 && bracketRight == len(currKey) -1 {
			index, error := strconv.Atoi(currKey[bracketLeft+1 : bracketRight])
			if error != nil || index < 0 {
				return nil, errors.Wrap(error, "key:"+" has index marker [], but the index is not valid integer.")
			} else if indexOccurance > 0 {
				return nil, errors.New("key:"+" has multiple index marker []. This is not supported")
			} else {
				currKey = currKey[0:bracketLeft]
				temp[currKey] = []interface{}{n}
				indexOccurance = indexOccurance + 1
			}
		} else {
			temp[currKey] = n
		}

		n = temp
	}
	return n, nil
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