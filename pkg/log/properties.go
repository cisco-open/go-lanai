package log

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"

const (
	defaultTemplate = `{{pad .time -25}} {{lvl . 5}} [{{pad .caller 25 | blue}}] {{pad .logger 12 | green}}: {{.msg}} {{kv .}}`
)

var defaultFixedFields = utils.NewStringSet(
	LogKeyMessage,
	LogKeyName,
	LogKeyTimestamp,
	LogKeyCaller,
	LogKeyLevel,
	LogKeyContext,
)

// Properties contains logging settings
// Note:
//	1. "context-mappings" indicate how to map context key to log key, it's map[context-key]log-key
type Properties struct {
	Levels   map[string]LoggingLevel     `json:"levels"`
	Loggers  map[string]LoggerProperties `json:"loggers"`
	Mappings map[string]string           `json:"context-mappings"`
}

// LoggerProperties individual logger setup
// Note:
//	1. we currently only support file and console type
//  2. "location" is ignored when "type" is "console"
// 	3. "template" and "fixed-keys" are ignored when "format" is not "text"
//	4. "template" is "text/template" compliant template, with "." as log KVs and following added functions:
//		- "{{padding .key -10}}" fixed length stringer
//		- "{{level . 5}}" colored level string with fixed length
//		- "{{coler .key}}" color code (red, green, yellow, gray, cyan) with pipeline support.
//			e.g. "{{padding .msg 20 | red}}"
type LoggerProperties struct {
	Type      LoggerType                `json:"type"`
	Format    Format                    `json:"format"`
	Location  string                    `json:"location"`
	Template  string                    `json:"template"`
	FixedKeys utils.CommaSeparatedSlice `json:"fixed-keys"`
}

func newProperties() *Properties {
	return &Properties{
		Levels: map[string]LoggingLevel{
			"default": LevelInfo,
		},
		Loggers:  map[string]LoggerProperties{
			"console": LoggerProperties{
				Type: TypeConsole,
				Format: FormatText,
				Template: defaultTemplate,
				FixedKeys: utils.CommaSeparatedSlice{
					LogKeyName, LogKeyMessage, LogKeyTimestamp,
					LogKeyCaller, LogKeyLevel, LogKeyContext,
				},
			},
		},
		Mappings: map[string]string{},
	}
}