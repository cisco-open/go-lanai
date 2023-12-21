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

package internal

import (
	"errors"
	"io"
)

var errMissingValue = errors.New("(MISSING)")

type Fields map[string]interface{}

// KitTextLoggerAdapter implmenets go-kit's log.Logger with custom Formatter
type KitTextLoggerAdapter struct {
	Formatter  TextFormatter
	Writer     io.Writer
	IsTerminal bool
}

func NewKitTextLoggerAdapter(writer io.Writer, formatter TextFormatter) *KitTextLoggerAdapter {
	return &KitTextLoggerAdapter{
		Formatter: formatter,
		Writer: writer,
		IsTerminal: IsTerminal(writer),
	}
}

func (l *KitTextLoggerAdapter) Log(keyvals ...interface{}) error {
	values := Fields{}
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			values[Sprint(keyvals[i])] = keyvals[i+1]
		} else {
			values[Sprint(keyvals[i])] = errMissingValue
		}
	}
	return l.Formatter.Format(values, l.Writer)
}


