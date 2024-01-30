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
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"time"
)

var zapBufferPool = buffer.NewPool()

// ZapFormattedEncoder implements zapcore.Encoder. This encoder leverage go template system for render user defined log.
// Note: Unlike zapcore's JSONEncoder and ConsoleEncoder, this encoder focus on flexibility rather than performance.
//		 When performance is crucial, JSON format of log should be used.
type ZapFormattedEncoder struct {
	*zapcore.MapObjectEncoder
	Formatter  TextFormatter
	Config     *zapcore.EncoderConfig
	IsTerminal bool
}

func NewZapFormattedEncoder(cfg zapcore.EncoderConfig, formatter TextFormatter, isTerm bool) zapcore.Encoder {
	return &ZapFormattedEncoder{
		MapObjectEncoder: zapcore.NewMapObjectEncoder(),
		Formatter:        formatter,
		Config:           &cfg,
		IsTerminal:       isTerm,
	}
}

func (enc *ZapFormattedEncoder) Clone() zapcore.Encoder {
	return &ZapFormattedEncoder{
		MapObjectEncoder: zapcore.NewMapObjectEncoder(),
		Formatter:        enc.Formatter,
		Config:           enc.Config,
		IsTerminal:       enc.IsTerminal,
	}
}

// EncodeEntry implements zapcore.Encoder
// We use map and slice based encoders. Working with map and slice is necessary with go template based formatter.
// Map and slice operations is not the most performant approach, but it wouldn't be the bottleneck comparing to go template rendering
func (enc *ZapFormattedEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	objEnc := zapcore.NewMapObjectEncoder()
	// encode entry
	arrEnc := &SliceArrayEncoder{make([]interface{}, 0, 5)}
	objEnc.Fields[enc.Config.NameKey] = applyZapValueEncoder(arrEnc, entry.LoggerName, enc.Config.EncodeName)
	objEnc.Fields[enc.Config.LevelKey] = applyZapValueEncoder(arrEnc, entry.Level, enc.Config.EncodeLevel)
	objEnc.Fields[enc.Config.TimeKey] = applyZapValueEncoder(arrEnc, entry.Time, enc.Config.EncodeTime)
	objEnc.Fields[enc.Config.MessageKey] = entry.Message
	if entry.Caller.Defined {
		objEnc.Fields[enc.Config.CallerKey] = applyZapValueEncoder(arrEnc, entry.Caller, enc.Config.EncodeCaller)
	}
	if len(entry.Stack) != 0 {
		objEnc.Fields[enc.Config.StacktraceKey] = entry.Stack
	}

	// encode fields
	for i := range fields {
		fields[i].AddTo(objEnc)
	}
	buf := zapBufferPool.Get()
	if e := enc.Formatter.Format(objEnc.Fields, buf); e != nil {
		return nil, e
	}
	return buf, nil
}

func applyZapValueEncoder[T any](arrEnc *SliceArrayEncoder, value T, valueEncoder func(T, zapcore.PrimitiveArrayEncoder)) interface{} {
	if valueEncoder == nil {
		return value
	}
	valueEncoder(value, arrEnc)
	return arrEnc.Latest()
}

// SliceArrayEncoder implementing zapcore.PrimitiveArrayEncoder. It's used to apply zapcore's entry encoders like zapcore.NameEncoder
type SliceArrayEncoder struct {
	elems []interface{}
}

func (s *SliceArrayEncoder) Latest() interface{} {
	return s.elems[len(s.elems)-1]
}

func (s *SliceArrayEncoder) AppendBool(v bool)              { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendByteString(v []byte)      { s.elems = append(s.elems, string(v)) }
func (s *SliceArrayEncoder) AppendComplex128(v complex128)  { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendComplex64(v complex64)    { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendDuration(v time.Duration) { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendFloat64(v float64)        { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendFloat32(v float32)        { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendInt(v int)                { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendInt64(v int64)            { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendInt32(v int32)            { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendInt16(v int16)            { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendInt8(v int8)              { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendString(v string)          { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendTime(v time.Time)         { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendUint(v uint)              { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendUint64(v uint64)          { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendUint32(v uint32)          { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendUint16(v uint16)          { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendUint8(v uint8)            { s.elems = append(s.elems, v) }
func (s *SliceArrayEncoder) AppendUintptr(v uintptr)        { s.elems = append(s.elems, v) }
