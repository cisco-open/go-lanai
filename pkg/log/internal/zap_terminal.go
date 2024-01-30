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
	"go.uber.org/zap/zapcore"
	"io"
)

type TerminalAware interface {
	IsTerminal() bool
}

// ZapWriterWrapper implements zapcore.WriteSyncer and TerminalAware
type ZapWriterWrapper struct {
	io.Writer
}

func (ZapWriterWrapper) Sync() error {
	return nil
}

func (s ZapWriterWrapper) IsTerminal() bool {
	return IsTerminal(s.Writer)
}

// NewZapWriterWrapper similar to zapcore.AddSync with exported type
func NewZapWriterWrapper(w io.Writer) zapcore.WriteSyncer {
	return ZapWriterWrapper{
		Writer: w,
	}
}

// ZapTerminalCore implements TerminalAware and always returns true
type ZapTerminalCore struct {
	zapcore.Core
}

func (s ZapTerminalCore) IsTerminal() bool {
	return true
}