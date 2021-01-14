package appconfig

import (
	"encoding/json"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"sort"
	"strconv"
	"strings"
)

var ErrNotLoaded = errors.New("Configuration not loaded")

type config struct {
	Providers     []Provider //such as yaml auth, commandline etc.
	settings      map[string]interface{}
	isLoaded 	  bool
}

type BootstrapConfig struct {
	*config
}

type ApplicationConfig struct {
	*config
}

type ConfigAccessor interface {
	Value(key string) interface{}
	Bind(target interface{}, prefix string) error
	Each(apply func(string, interface{}) error) error
}

func NewBootstrapConfig(providers ...Provider) *BootstrapConfig {
	return &BootstrapConfig{&config{Providers: providers}}
}

func NewApplicationConfig(providers ...Provider) *ApplicationConfig {
	return &ApplicationConfig{&config{Providers: providers}}
}

//load will fail if place holder cannot be resolved due to circular dependency
func (c *config) Load(force bool) (loadError error) {
	defer func() {
		if loadError != nil {
			c.isLoaded = false
		} else {
			c.isLoaded = true
		}
	}()

	//sort based on precedence
	sort.SliceStable(c.Providers, func(i, j int) bool { return c.Providers[i].GetPrecedence() > c.Providers[j].GetPrecedence() })

	// Load appconfig from each auth if it's not loaded yet, or if force reload.
	for _, provider := range c.Providers {
		if !provider.isLoaded() || force {
			error := provider.Load()

			if error != nil {
				return errors.Wrap(error, "Failed to load properties")
			}
		}
	}

	merged := make(map[string]interface{})
	// merge data
	mergeOption := func(mergoConfig *mergo.Config) {
		mergoConfig.Overwrite = true
	}

	for _, provider := range c.Providers {
		if provider.GetSettings() != nil {
			formatted, error := ProcessKeyFormat(provider.GetSettings(), NormalizeKey)
			if error != nil {
				return errors.Wrap(error, "Failed to format keys before merge")
			}
			mergeError := mergo.Merge(&merged, formatted, mergeOption)
			if mergeError != nil {
				return errors.Wrap(error, "Failed to merge properties from property sources")
			}
		}
	}

	error := resolve(merged)

	if error != nil {
		return error
	}

	c.settings = merged
	return nil
}

func (c *config) Value(key string) interface{} {
	if !c.isLoaded {
		return nil
	}

	return value(c.settings, key)
}

func value(nested map[string]interface{}, flatKey string) interface{} {
	targetKey := NormalizeKey(flatKey)

	nestedKeys := UnFlattenKey(targetKey)

	var tmp interface{} = nested
	for i, nestedKey := range nestedKeys {
		indexStart := strings.Index(nestedKey, "[")
		indexEnd := strings.Index(nestedKey, "]")

		var index int = -1
		if indexStart > -1 && indexEnd > -1 {
			indexStr := nestedKey[indexStart+1 : indexEnd]
			index, _ = strconv.Atoi(indexStr)
			nestedKey = nestedKey[0:indexStart]
		}

		//if we are not at the leaf yet
		//we move tmp to the next level
		if i < len(nestedKeys)-1 {
			tmp = (tmp.(map[string]interface{}))[nestedKey]
			if index > -1  {
				tmp = tmp.([]interface{})[index]
			}
		} else {
			if index > -1 {
				tmp = ((tmp.(map[string]interface{}))[nestedKey].([]interface{}))[index]
			} else {
				tmp = (tmp.(map[string]interface{}))[nestedKey]
			}
		}
	}
	return tmp
}

/*
 * The keys from property sources are normalized to snake case if they are camel case.
 * Therefore the binding expects the json tag to be in snake case.
 */
func (c *config) Bind(target interface{}, prefix string) error {
	keys := strings.Split(prefix, ".")

	var source interface{} = c.settings

	for i := 0; i < len(keys); i++ {

		if _, ok := source.(map[string] interface{}); ok {
			source = source.(map[string] interface{})[keys[i]]
		} else {
			return errors.New("prefix doesn't exist")
		}
	}

	serialized, error := json.Marshal(source)

	if error == nil {
		error = json.Unmarshal(serialized, target)
	}

	return error
}

//stops at the first error
func (c *config) Each(apply func(string, interface{})error) error{
	return VisitEach(c.settings, apply)
}

func resolve(nested map[string]interface{}) error {
	doResolve := func(key string, value interface{}) error {
		_, error := resolveValue(nested, key, key)
		return error
	}

	error := VisitEach(nested, doResolve)
	if error != nil {
		return error
	}



	return nil
}

//here the key is the flattened key
func resolveValue(source map[string]interface{}, key string, originKey string) (interface{}, error) {
	value := value(source, key)

	//if value is not string, no need to resolve it further
	if _, ok := value.(string); !ok {
		return value, nil
	}

	placeHolderKeys, isEmbedded, error := parsePlaceHolder(value.(string))

	if error != nil {
		return "", error
	}

	//There's no place holder in the value
	//return the value
	if len(placeHolderKeys) == 0 {
		return value, nil
	}

	fmt.Println("resolving key: " + key)
	for _, placeHolderKey := range placeHolderKeys {
		if strings.Compare(originKey, placeHolderKey) == 0 {
			return "", errors.New("key: " + originKey + " can't be resolved due to circular reference")
		}
	}

	resolvedKV := make(map[string]interface{})
	for _, placeHolderKey := range placeHolderKeys {
		resolvedPlaceHolder, error := resolveValue(source, placeHolderKey, originKey)
		if error != nil {
			return "", error
		}
		resolvedKV[placeHolderKey] = resolvedPlaceHolder
	}

	//embedded means the place holder is embedded in the value string, either with the format of
	// somestring${a}
	// or
	// ${a}${b}
	//therefore the resolved place holders must be all strings as well, otherwise we can't concatenate them together.
 	if isEmbedded {
		resolvedValue := value.(string)
		for placeHolderKey, resolvedPlaceHolder := range resolvedKV {
			resolvedValue = strings.Replace(resolvedValue, placeHolderPrefix+placeHolderKey+placeHolderSuffix, fmt.Sprint(resolvedPlaceHolder), -1)
		}
		updateMapUsingFlatKey(source, key, resolvedValue)
		return resolvedValue, nil
	 } else { //if not embedded, the entire value must have just been a single place holder.
		resolvedValue := resolvedKV[placeHolderKeys[0]]
		updateMapUsingFlatKey(source, key, resolvedValue)
		return resolvedValue, nil
	}
}

func updateMapUsingFlatKey(source map[string]interface{}, flatKey string, value interface{}) {
	nestedKeys := UnFlattenKey(flatKey)

	var tmp interface{} = source
	for i, nestedKey := range nestedKeys {
		//TODO: depulicate this index logic
		indexStart := strings.Index(nestedKey, "[")
		indexEnd := strings.Index(nestedKey, "]")

		var index int = -1
		if indexStart > -1 && indexEnd > -1 {
			indexStr := nestedKey[indexStart+1 : indexEnd]
			index, _ = strconv.Atoi(indexStr)
			nestedKey = nestedKey[0:indexStart]
		}

		//if we are not at the leaf yet
		//we move tmp to the next level
		if i < len(nestedKeys)-1 {
			tmp = (tmp.(map[string]interface{}))[nestedKey]
			if index > -1  {
				tmp = tmp.([]interface{})[index]
			}
		} else {
			if index > -1 {
				((tmp.(map[string]interface{}))[nestedKey].([]interface{}))[index] = value
			} else {
				(tmp.(map[string]interface{}))[nestedKey] = value
			}
		}
	}
}

const placeHolderPrefix = "${"
const placeHolderSuffix = "}"

type bracket struct {
	value string
	index int
}

func parsePlaceHolder(strValue string) (placeHolderKeys []string, isEmbedded bool, error error) {
	//use this as a stack to check for nested place holder brackets
	//the algorithm is to put left bracket on the stack, and pop it off when we see a right bracket
	//this way if the stack is at length greater than 1 when we encounter another left bracket, we have a nested situation
	var bracketStack []bracket

	for i := 0; i < len(strValue); i++ {
		//if we encounters ${
		if i <= len(strValue) - len(placeHolderPrefix) && strings.Compare(strValue[i:i+len(placeHolderPrefix)], placeHolderPrefix) == 0 {
			bracketStack = append(bracketStack, bracket{placeHolderPrefix, i+1})
			if len(bracketStack) > 1 {
				return nil, false, errors.New(strValue + " has nested place holders, which is not supported")
			}
		}

		//if we encounter }
		if strings.Compare(strValue[i:i+1], placeHolderSuffix) == 0 {
			stackLen := len(bracketStack)
			if stackLen >= 1 {
				leftBracket := bracketStack[stackLen -1] //gets the top of the stack
				bracketStack = bracketStack[:stackLen-1] //pop the top of the stack
				placeHolderKeys = append(placeHolderKeys, strValue[leftBracket.index + 1 : i])

				if leftBracket.index > len(placeHolderPrefix) -1 || i < len(strValue) - 1 {
					isEmbedded = true
				}
			} //else there's no matching ${, so we skip it
		}
	}
	return placeHolderKeys, isEmbedded, nil
}
