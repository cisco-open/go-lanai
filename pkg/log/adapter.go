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

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"strings"
)

// writerAdapter implements io.Writer and wrap around our Logger interface
type writerAdapter struct {
	logger log.Logger
}

func NewWriterAdapter(logger Logger, lvl LoggingLevel) *writerAdapter {
	kitLogger := log.Logger(logger)
	switch lvl {
	case LevelDebug:
		kitLogger = level.Debug(logger)
	case LevelInfo:
		kitLogger = level.Info(logger)
	case LevelWarn:
		kitLogger = level.Warn(logger)
	case LevelError:
		kitLogger = level.Error(logger)
	default:
		// TODO should be a noop kit logger
	}
	return &writerAdapter{
		logger: kitLogger,
	}
}

func (w writerAdapter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	msg := strings.TrimSpace(string(p))
	return len(p), w.logger.Log(LogKeyMessage, msg)
}
