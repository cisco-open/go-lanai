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

package httpclient

import (
    "embed"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/pkg/errors"
    "strings"
    "time"
)

const (
	PropertiesPrefix = "integrate.http"
)

const (
	LogDetailsLevelUnknown LogDetailsLevel = iota
	LogDetailsLevelNone
	LogDetailsLevelMinimum
	LogDetailsLevelHeaders
	LogDetailsLevelFull
)

const (
	logDetailsLevelUnknownText = "unknown"
	logDetailsLevelNoneText    = "off"
	logDetailsLevelMinimumText = "minimum"
	logDetailsLevelHeadersText = "headers"
	logDetailsLevelFullText    = "full"
)

var (
	logDetailsLevelAtoI = map[string]LogDetailsLevel{
		strings.ToLower(logDetailsLevelUnknownText): LogDetailsLevelUnknown,
		strings.ToLower(logDetailsLevelNoneText):    LogDetailsLevelNone,
		strings.ToLower(logDetailsLevelMinimumText): LogDetailsLevelMinimum,
		strings.ToLower(logDetailsLevelHeadersText): LogDetailsLevelHeaders,
		strings.ToLower(logDetailsLevelFullText):    LogDetailsLevelFull,
	}

	logDetailsLevelItoA = map[LogDetailsLevel]string{
		LogDetailsLevelUnknown: logDetailsLevelUnknownText,
		LogDetailsLevelNone:    logDetailsLevelNoneText,
		LogDetailsLevelMinimum: logDetailsLevelMinimumText,
		LogDetailsLevelHeaders: logDetailsLevelHeadersText,
		LogDetailsLevelFull:    logDetailsLevelFullText,
	}
)

type LogDetailsLevel int

func (l LogDetailsLevel) String() string {
	if s, ok := logDetailsLevelItoA[l]; ok {
		return s
	}
	return logDetailsLevelNoneText
}

func (l LogDetailsLevel) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l *LogDetailsLevel) UnmarshalText(data []byte) error {
	value := strings.ToLower(string(data))
	if v, ok := logDetailsLevelAtoI[value]; ok {
		*l = v
	}
	return nil
}

//go:embed defaults-integrate-http.yml
var defaultConfigFS embed.FS

type HttpClientProperties struct {
	MaxRetries int              `json:"max-retries"` // negative value means no retry
	Timeout    utils.Duration   `json:"timeout"`
	Logger     LoggerProperties `json:"logger"`
}

type LoggerProperties struct {
	Level           log.LoggingLevel          `json:"level"`
	DetailsLevel    LogDetailsLevel           `json:"details-level"`
	SanitizeHeaders utils.CommaSeparatedSlice `json:"sanitize-headers"`
	ExcludeHeaders  utils.CommaSeparatedSlice `json:"exclude-headers"`
}

func newHttpClientProperties() *HttpClientProperties {
	return &HttpClientProperties{
		MaxRetries: 3,
		Timeout:    utils.Duration(1 * time.Minute),
		Logger: LoggerProperties{
			Level:           log.LevelDebug,
			DetailsLevel:    LogDetailsLevelHeaders,
			SanitizeHeaders: utils.CommaSeparatedSlice{HeaderAuthorization},
			ExcludeHeaders: utils.CommaSeparatedSlice{},
		},
	}
}

func bindHttpClientProperties(ctx *bootstrap.ApplicationContext) HttpClientProperties {
	props := newHttpClientProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind HttpClientProperties"))
	}
	return *props
}
