package appconfig

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"encoding/json"
	"fmt"
	"dario.cat/mergo"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

var (
	logger = log.New("Config")
	//ErrNotLoaded = errors.New("Configuration not loaded")
	errBindWithConfigBeforeLoaded = errors.New("attempt to bind with config before it's loaded")
)

// properties implements bootstrap.ApplicationConfig
type properties map[string]interface{}

func makeInitialProperties() properties {
	return map[string]interface{}{}
}

func (p properties) Value(key string) interface{} {
	return value(p, key)
}

// Bind bind values to given target, with consideration of key normalization and place holders.
// The keys from property sources are normalized to snake case if they are camel case.
// Therefore, the binding expects the json tag to be in snake case.
func (p properties) Bind(target interface{}, prefix string) error {
	var source interface{} = map[string]interface{}(p)
	if len(prefix) > 0 && prefix != "." {
		keys := strings.Split(prefix, ".")
		for i := 0; i < len(keys); i++ {
			if _, ok := source.(map[string]interface{}); ok {
				source = source.(map[string]interface{})[keys[i]]
			} else {
				//prefix doesn't exist, we just don't bind it
				return nil
			}
		}
	}

	serialized, e := json.Marshal(source)

	if e == nil {
		e = json.Unmarshal(serialized, target)
	}

	return e
}

type config struct {
	properties
	groups    []ProviderGroup
	providers []Provider //such as yaml auth, commandline etc.
	profiles  utils.StringSet
	isLoaded  bool
}

//Load will fail if place holder cannot be resolved due to circular dependency
func (c *config) Load(ctx context.Context, force bool) (err error) {
	defer func() {
		if err != nil {
			c.isLoaded = false
		} else {
			c.isLoaded = true
		}
	}()

	// sort groups based on order, we process lower priority first
	order.SortStable(c.groups, order.OrderedFirstCompareReverse)

	// reset all groups if force == true
	if force {
		c.isLoaded = false
		c.profiles = nil
		for _, g := range c.groups {
			g.Reset()
		}
	}

	// repeatedly process provider groups until list of provider become stable and all loaded
	var providers []Provider
	final := makeInitialProperties()
	// Note about hasNew check: when transiting from bootstrap config to application config,
	// and all initial providers are from bootstrap config, all providers are loaded initially.
	// However, we still need to re-collect/merge all properties.
	// In this case, we need to set hasNew to true in the first iteration.
	for hasNew, isFirstIter := true, true; hasNew; {
		providers = make([]Provider, 0)
		hasNew = isFirstIter
		isFirstIter = false
		for _, g := range c.groups {
			// sort providers based on precedence, lower to higher
			group := g.Providers(ctx, final)
			order.SortStable(group, order.OrderedFirstCompareReverse)
			providers = append(providers, group...)

			// Load config from each source if it's not loaded yet
			for _, provider := range group {
				if !provider.IsLoaded() {
					if e := provider.Load(ctx); e != nil {
						err = errors.Wrap(e, "Failed to load properties")
						return
					}
					hasNew = true
				}
			}
		}

		// If no new provider are loaded we quick without re-merge all sources
		if !hasNew {
			break
		}

		// merge properties and deal special merging rules on some properties
		// Note all properties returned by Provider should be un-flattened
		merged := makeInitialProperties()
		additionalProfiles := make([]string, 0)
		for _, p := range providers {
			if p.GetSettings() == nil {
				continue
			}

			formatted, e := ProcessKeyFormat(p.GetSettings(), NormalizeKey)
			if e != nil {
				err = errors.Wrap(e, "Failed to format keys before merge")
				return
			}

			if e := mergo.Merge(&merged, properties(formatted), mergo.WithOverride); e != nil {
				err = errors.Wrap(e, "Failed to merge properties from property sources")
				return
			}

			// special treatments:
			// 	- PropertyKeyAdditionalProfiles need to be appended instead of overridden
			if additionalProfiles, e = mergeAdditionalProfiles(additionalProfiles, formatted); e != nil {
				return e
			}
		}

		if e := setValue(merged, PropertyKeyAdditionalProfiles, additionalProfiles, true); e != nil {
			return e
		}
		final = merged
	}

	// resolve placeholder
	if err = resolve(ctx, final); err != nil {
		return
	}
	c.properties = final

	// resolve profiles
	c.profiles = utils.NewStringSet()
	for _, v := range resolveProfiles(final) {
		if v != "" {
			c.profiles.Add(v)
		}
	}

	// providers are stored in highest precedence first
	l := len(providers)
	c.providers = make([]Provider, l)
	for i, v := range providers {
		c.providers[l-i-1] = v
	}

	return
}

func (c *config) Value(key string) interface{} {
	if !c.isLoaded {
		return nil
	}

	return c.properties.Value(key)
}

func (c *config) Bind(target interface{}, prefix string) error {
	if !c.isLoaded {
		return errBindWithConfigBeforeLoaded
	}
	return c.properties.Bind(target, prefix)
}

// Each go through all properties and apply given function.
// It stops at the first error
func (c *config) Each(apply func(string, interface{}) error) error {
	return VisitEach(c.properties, apply)
}

func (c *config) Providers() []Provider {
	return c.providers
}

func (c *config) Profiles() []string {
	return c.profiles.Values()
}

func (c *config) HasProfile(profile string) bool {
	return c.profiles.Has(profile)
}

func (c *config) ProviderGroups() []ProviderGroup {
	return c.groups
}

/*********************
	Helpers
 *********************/
func value(nested map[string]interface{}, flatKey string) (ret interface{}) {
	e := visit(nested, flatKey, func(_ string, v interface{}, isLeaf bool, _ int) interface{} {
		if isLeaf {
			ret = v
		}
		return nil
	})
	if e != nil {
		return nil
	}
	return
}

// setValue set given val in map using flat key.
// Note: this method won't create intermediate node if it already exist.
//		 This means out of bound error is still possible
func setValue(nested map[string]interface{}, flatKey string, val interface{}, createIntermediateNodes bool) error {
	return visit(nested, flatKey, func(_ string, v interface{}, isLeaf bool, expectedSliceLen int) interface{} {
		switch {
		case isLeaf:
			return val
		case v != nil || !createIntermediateNodes:
			// non leaf item, we do nothing if existing value is not nil
			return nil
		case expectedSliceLen > 0:
			// create intermediate slice
			s := make([]interface{}, expectedSliceLen)
			for i := 0; i < expectedSliceLen; i++ {
				s[i] = map[string]interface{}{}
			}
			return s
		default:
			// create intermediate map
			return map[string]interface{}{}
		}
	})
}

type visitFunc func(keyPath string, v interface{}, isLeaf bool, expectedSliceLen int) interface{}

// visit traverse the given tree (map) along the path represented as flatKey (e.g. flat.key[0].path)
// it calls overrideFunc with each node's partial key path and its value. if returned value is non-nil, it will
// replace the node
func visit(nested map[string]interface{}, flatKey string, overrideFunc visitFunc) error  {
	targetKey := NormalizeKey(flatKey)
	nestedKeys := UnFlattenKey(targetKey)
	partialKey := ""
	var tmp interface{} = nested
	for i, nestedKey := range nestedKeys {
		// set index if nested key is "[index]" format
		var index = -1
		indexStart := strings.Index(nestedKey, "[")
		indexEnd := strings.Index(nestedKey, "]")
		if indexStart > -1 && indexEnd > -1 {
			indexStr := nestedKey[indexStart+1 : indexEnd]
			index, _ = strconv.Atoi(indexStr)
			nestedKey = nestedKey[0:indexStart]
		}

		m, ok := tmp.(map[string]interface{})
		if !ok {
			return fmt.Errorf("incorrect type at key path %s. expected map[string]interface{}, but got %T", partialKey, tmp)
		}

		// get value and attempt to override
		tmp = m[nestedKey]
		isLast := i == len(nestedKeys) - 1
		partialKey = joinKeyPaths(partialKey, nestedKey)
		if v := overrideFunc(partialKey, tmp, isLast && index < 0, index + 1); v != nil {
			m[nestedKey] = v
			tmp = v
		}

		if index >= 0 {
			// slice
			s, ok := tmp.([]interface{})
			if !ok || len(s) <= index {
				return fmt.Errorf("index %d out of bound (%d) at key path %s", index, len(s), partialKey)
			}
			// attempt to override
			tmp = s[index]
			partialKey = joinKeyPaths(partialKey, nestedKey)
			if v := overrideFunc(partialKey, tmp, isLast, -1); v != nil {
				s[index] = v
				tmp = v
			}
		}
	}
	return nil
}

func joinKeyPaths(left string, right interface{}) string {
	switch r := right.(type) {
	case string:
		switch {
		case left == "":
			return r
		case right == "":
			return left
		default:
			return left + "." + r
		}
	case int:
		return left + "[" + strconv.Itoa(r) + "]"
	default:
		return ""
	}
}

func mergeAdditionalProfiles(profiles []string, src map[string]interface{}) ([]string, error) {
	raw := value(src, PropertyKeyAdditionalProfiles)
	switch v := raw.(type) {
	case nil:
		return profiles, nil
	case string:
		profiles = append(profiles, v)
	case []string:
		profiles = append(profiles, v...)
	case []interface{}:
		for i, p := range v {
			s, ok := p.(string)
			if !ok {
				return nil, fmt.Errorf("invalid type %T at key path %s[%d]", v, PropertyKeyAdditionalProfiles, i)
			}
			profiles = append(profiles, s)
		}
	default:
		return nil, fmt.Errorf("invalid type %T at key path %s", raw, PropertyKeyAdditionalProfiles)
	}
	return profiles, nil
}

/*********************
	Placeholder
 *********************/

func resolve(ctx context.Context, nested map[string]interface{}) error {
	doResolve := func(key string, value interface{}) error {
		_, e := resolveValue(ctx, key, value, nested, nil)
		return e
	}

	if e := VisitEach(nested, doResolve); e != nil {
		return e
	}
	return nil
}

// resolveValue recursively resolve the value of key by replacing placeholders with actual value
// Note: here the key is the flattened key
func resolveValue(ctx context.Context, key string, val interface{}, source map[string]interface{}, visited []string) (resolvedVal interface{}, err error) {
	//if value is not string, no need to resolve it further
	if _, ok := val.(string); !ok {
		return val, nil
	}

	placeholders, isEmbedded, e := parsePlaceHolder(val.(string))
	if e != nil {
		return "", e
	} else if len(placeholders) == 0 {
		return val, nil
	}

	// check for circular reference
	visited = append(visited, key)
	logger.WithContext(ctx).Debugf("resolving key: " + key)
	for _, ph := range placeholders {
		for i, k := range visited {
			if strings.Compare(k, ph.key) == 0 {
				circular := strings.Join(visited[i:], "->") + "->" + ph.key
				return "", fmt.Errorf("placeholder ${%s} can't be resolved due to circular reference: %s", ph.key, circular)
			}
		}
	}

	resolvedKV := make(map[string]interface{})
	resolvedPlaceholder := make(map[string]placeholder)
	for _, ph := range placeholders {
		v := value(source, ph.key)
		switch resolved, e := resolveValue(ctx, ph.key, v, source, visited); {
		case e == nil && resolved == nil:
			// cannot resolve value
			if ph.defaultVal != nil {
				resolvedKV[ph.key] = ph.defaultVal
				resolvedPlaceholder[ph.key] = ph
			}
		case e == nil:
			resolvedKV[ph.key] = resolved
			resolvedPlaceholder[ph.key] = ph
		case e != nil && ph.defaultVal != nil:
			logger.WithContext(ctx).Warnf(e.Error())
			resolvedKV[ph.key] = ph.defaultVal
			resolvedPlaceholder[ph.key] = ph
		default:
			return nil, e
		}
	}

	// embedded means the placeholder is embedded in the value string, either with the format of
	// "somestring${a}" or "${a}${b}"
	// therefore the resolvedVal placeholders must be all strings as well, otherwise we can't concatenate them together.
	var resolvedValue interface{}
	if isEmbedded {
		str := val.(string)
		for phKey, resolved := range resolvedKV {
			str = strings.Replace(str, resolvedPlaceholder[phKey].String(), fmt.Sprint(resolved), -1)
		}
		resolvedValue = str
	} else { //if not embedded, the entire value must have just been a single placeholder.
		resolvedValue = resolvedKV[placeholders[0].key]
	}

	if e := setValue(source, key, resolvedValue, false); e != nil {
		return nil, e
	}

	return resolvedValue, nil
}

const placeHolderPrefix = "${"
const placeHolderSuffix = "}"
const placeHolderDefaultDelimiter = ":"

type bracket struct {
	value string
	index int
}

type placeholder struct {
	key        string
	defaultVal interface{}
}

func (ph placeholder) String() string {
	if ph.defaultVal == nil {
		return fmt.Sprintf("%s%s%s", placeHolderPrefix, ph.key, placeHolderSuffix)
	}
	return fmt.Sprintf("%s%s%s%v%s", placeHolderPrefix, ph.key, placeHolderDefaultDelimiter, ph.defaultVal, placeHolderSuffix)
}

// embedded means the placeholder is embedded in the value string, either with the format of "somestring${a}" or "${a}${b}"
// Note: when default value is present e.g. "${non-exist.key:default_value}", the type of default value is unknown
//		 (information about whether the value was quoted is lost during YAML parsing).
//		 We guess the type based on the default value using strconv package:
// 			- 100 -> json.Number
// 			- 100.0 -> float
// 			- true/false -> bool
// 			- other values -> string
func parsePlaceHolder(strValue string) (placeholders []placeholder, isEmbedded bool, error error) {
	//use this as a stack to check for nested placeholder brackets
	//the algorithm is to put left bracket on the stack, and pop it off when we see a right bracket
	//this way if the stack is at length greater than 1 when we encounter another left bracket, we have a nested situation
	var bracketStack []bracket

	for i := 0; i < len(strValue); i++ {
		//if we encounters ${
		if i <= len(strValue)-len(placeHolderPrefix) && strings.Compare(strValue[i:i+len(placeHolderPrefix)], placeHolderPrefix) == 0 {
			bracketStack = append(bracketStack, bracket{placeHolderPrefix, i + 1})
			if len(bracketStack) > 1 {
				return nil, false, errors.New(strValue + " has nested place holders, which is not supported")
			}
		}

		//if we encounter }
		if strings.Compare(strValue[i:i+1], placeHolderSuffix) == 0 {
			stackLen := len(bracketStack)
			if bracketStack != nil && stackLen >= 1 {
				leftBracket := bracketStack[stackLen-1]  //gets the top of the stack
				bracketStack = bracketStack[:stackLen-1] //pop the top of the stack
				split := strings.SplitN(strValue[leftBracket.index+1 : i], placeHolderDefaultDelimiter, 2)
				ph := placeholder{
					key:        split[0],
				}
				if len(split) > 1 {
					ph.defaultVal = utils.ParseString(split[1])
				}
				placeholders = append(placeholders, ph)

				if leftBracket.index > len(placeHolderPrefix)-1 || i < len(strValue)-1 {
					isEmbedded = true
				}
			} //else there's no matching ${, so we skip it
		}
	}
	return placeholders, isEmbedded, nil
}
