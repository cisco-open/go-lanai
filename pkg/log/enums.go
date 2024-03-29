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

package log

import "strings"

/*********************
	LoggingLevel
 *********************/

type LoggingLevel int

const (
	LevelOff LoggingLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

const (
	LevelOffText   = "OFF"
	LevelDebugText = "DEBUG"
	LevelInfoText  = "INFO"
	LevelWarnText  = "WARN"
	LevelErrorText = "ERROR"
)

var (
	loggingLevelAtoI = map[string]LoggingLevel{
		strings.ToUpper(LevelOffText):   LevelOff,
		strings.ToUpper(LevelDebugText): LevelDebug,
		strings.ToUpper(LevelInfoText):  LevelInfo,
		strings.ToUpper(LevelWarnText):  LevelWarn,
		strings.ToUpper(LevelErrorText): LevelError,
	}

	loggingLevelItoA = map[LoggingLevel]string{
		LevelOff:   LevelOffText,
		LevelDebug: LevelDebugText,
		LevelInfo:  LevelInfoText,
		LevelWarn:  LevelWarnText,
		LevelError: LevelErrorText,
	}
)

// String implements fmt.Stringer
func (l LoggingLevel) String() string {
	if s, ok := loggingLevelItoA[l]; ok {
		return s
	}
	return LevelErrorText
}

// MarshalText implements encoding.TextMarshaler
func (l LoggingLevel) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (l *LoggingLevel) UnmarshalText(data []byte) error {
	value := strings.ToUpper(string(data))
	if v, ok := loggingLevelAtoI[value]; ok {
		*l = v
	}
	return nil
}

/*********************
	Format
 *********************/

type Format int

const (
	_ Format = iota
	FormatText
	FormatJson
)

const (
	FormatJsonText = "json"
	FormatTextText = "text"
)

var (
	formatAtoI = map[string]Format{
		FormatJsonText: FormatJson,
		FormatTextText: FormatText,
	}

	formatItoA = map[Format]string{
		FormatJson: FormatJsonText,
		FormatText: FormatTextText,
	}
)

// fmt.Stringer
func (f Format) String() string {
	if s, ok := formatItoA[f]; ok {
		return s
	}
	return "unknown"
}

// encoding.TextMarshaler
func (f Format) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

// encoding.TextUnmarshaler
func (f *Format) UnmarshalText(data []byte) error {
	value := strings.ToLower(string(data))
	if v, ok := formatAtoI[value]; ok {
		*f = v
	}
	return nil
}

/*********************
	LoggerType
 *********************/
type LoggerType int

const (
	_ LoggerType = iota
	TypeConsole
	TypeFile
	TypeHttp
	TypeMQ
)

const (
	TypeConsoleText = "console"
	TypeFileText = "file"
	TypeHttpText = "http"
	TypeMQText = "mq"
)

var (
	typeAtoI = map[string]LoggerType{
		TypeConsoleText: TypeConsole,
		TypeFileText:    TypeFile,
		TypeHttpText:    TypeHttp,
		TypeMQText:      TypeMQ,
	}

	typeItoA = map[LoggerType]string{
		TypeConsole: TypeConsoleText,
		TypeFile: TypeFileText,
		TypeHttp: TypeHttpText,
		TypeMQ: TypeMQText,
	}
)

// fmt.Stringer
func (t LoggerType) String() string {
	if s, ok := typeItoA[t]; ok {
		return s
	}
	return "unknown"
}

// encoding.TextMarshaler
func (t LoggerType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// encoding.TextUnmarshaler
func (t *LoggerType) UnmarshalText(data []byte) error {
	value := strings.ToLower(string(data))
	if v, ok := typeAtoI[value]; ok {
		*t = v
	}
	return nil
}
