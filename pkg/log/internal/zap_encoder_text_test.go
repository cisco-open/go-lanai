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
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	"testing"
	"time"
)

/*************************
	Tests
 *************************/

func TestZapFormattedEncoder(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestClone(), "Clone"),
	)
}

func TestSliceArrayEncoder(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestAppend(), "Append"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestClone() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		enc := NewZapFormattedEncoder(zapcore.EncoderConfig{}, nil, true).(*ZapFormattedEncoder)
		enc.AddBool("whatever", true)
		clone := enc.Clone().(*ZapFormattedEncoder)
		g.Expect(clone.MapObjectEncoder.Fields).To(gomega.BeEmpty())
		//g.Expect(clone.Formatter).To(gomega.Equal(enc.Formatter))
		g.Expect(clone.Config).To(gomega.Equal(enc.Config))
		g.Expect(clone.IsTerminal).To(gomega.Equal(enc.IsTerminal))
	}
}

func SubTestAppend() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		enc := &SliceArrayEncoder{make([]interface{}, 0, 20)}
		AssertAppend(g, enc, true, enc.AppendBool, true)
		AssertAppend(g, enc, []byte("test"), enc.AppendByteString, "test")
		AssertAppend(g, enc, 1+1i, enc.AppendComplex128, 1+1i)
		AssertAppend(g, enc, 1+1i, enc.AppendComplex64, complex64(1+1i))
		AssertAppend(g, enc, time.Second, enc.AppendDuration, time.Second)
		AssertAppend(g, enc, 1.1, enc.AppendFloat64, 1.1)
		AssertAppend(g, enc, 1.1, enc.AppendFloat32, float32(1.1))
		AssertAppend(g, enc, -100, enc.AppendInt, -100)
		AssertAppend(g, enc, int64(-100), enc.AppendInt64, int64(-100))
		AssertAppend(g, enc, int32(-100), enc.AppendInt32, int32(-100))
		AssertAppend(g, enc, int16(-100), enc.AppendInt16, int16(-100))
		AssertAppend(g, enc, int8(-100), enc.AppendInt8, int8(-100))
		AssertAppend(g, enc, "test-string", enc.AppendString, "test-string")
		now := time.Now()
		AssertAppend(g, enc, now, enc.AppendTime, now)
		AssertAppend(g, enc, uint(100), enc.AppendUint, uint(100))
		AssertAppend(g, enc, uint64(100), enc.AppendUint64, uint64(100))
		AssertAppend(g, enc, uint32(100), enc.AppendUint32, uint32(100))
		AssertAppend(g, enc, uint16(100), enc.AppendUint16, uint16(100))
		AssertAppend(g, enc, uint8(100), enc.AppendUint8, uint8(100))
		AssertAppend(g, enc, uintptr(100), enc.AppendUintptr, uintptr(100))
	}
}

/*************************
	Helpers
 *************************/

func AssertAppend[T any](g *gomega.WithT, enc *SliceArrayEncoder, v T, appender func(T), expected interface{}) {
	appender(v)
	g.Expect(enc.Latest()).To(gomega.Equal(expected), "appending %T should not fail", v)
}