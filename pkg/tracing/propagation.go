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

package tracing

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/zipkin"
	"strconv"
	"strings"
)

// Option is a function that sets an option on Propagator
type Option func(propagator *Propagator)

// BaggagePrefix is a function that sets baggage prefix on Propagator
//goland:noinspection GoUnusedExportedFunction
func BaggagePrefix(prefix string) Option {
	return func(propagator *Propagator) {
		propagator.baggagePrefix = prefix
		zipkin.BaggagePrefix(prefix)(&propagator.delegate)
	}
}

func SingleHeader() Option {
	return func(propagator *Propagator) {
		propagator.singleHeader = true
	}
}

// Propagator is an extension of zipkin.Propagator that support Single Header propagation:
// See https://github.com/openzipkin/b3-propagation#single-header
type Propagator struct {
	delegate      zipkin.Propagator
	baggagePrefix string
	singleHeader  bool
}

// NewZipkinB3Propagator creates a Propagator for extracting and injecting Zipkin B3 headers into SpanContexts.
// Baggage is by default enabled and uses prefix 'baggage-'.
func NewZipkinB3Propagator(opts ...Option) *Propagator {
	p := Propagator{
		delegate: zipkin.NewZipkinB3HTTPHeaderPropagator(),
	}
	for _, fn := range opts {
		fn(&p)
	}
	return &p
}

// Inject conforms to the Injector interface for decoding Zipkin B3 headers
func (p Propagator) Inject(sc jaeger.SpanContext, abstractCarrier interface{}) error {
	if !p.singleHeader {
		return p.delegate.Inject(sc, abstractCarrier)
	}

	// single header
	textMapWriter, ok := abstractCarrier.(opentracing.TextMapWriter)
	if !ok {
		return opentracing.ErrInvalidCarrier
	}

	// https://github.com/openzipkin/b3-propagation#single-header
	// b3={TraceId}-{SpanId}-{SamplingState}-{ParentSpanId}, where the last two fields are optional.
	values := []string{
		sc.TraceID().String(),
		sc.SpanID().String(),
	}

	if sc.IsSampled() {
		values = append(values, "1")
	} else {
		values = append(values, "0")
	}

	if sc.ParentID() != 0 {
		values = append(values, strconv.FormatUint(uint64(sc.ParentID()), 16))
	}

	textMapWriter.Set("b3", strings.Join(values, "-"))

	sc.ForeachBaggageItem(func(k, v string) bool {
		textMapWriter.Set(p.baggagePrefix+k, v)
		return true
	})
	return nil
}

// Extract conforms to the Extractor interface for encoding Zipkin HTTP B3 headers
func (p Propagator) Extract(abstractCarrier interface{}) (jaeger.SpanContext, error) {
	if !p.singleHeader {
		return p.delegate.Extract(abstractCarrier)
	}

	textMapReader, ok := abstractCarrier.(opentracing.TextMapReader)
	if !ok {
		return jaeger.SpanContext{}, opentracing.ErrInvalidCarrier
	}

	var traceID jaeger.TraceID
	var spanID jaeger.SpanID
	var parentID uint64
	sampled := false

	var baggage map[string]string
	err := textMapReader.ForeachKey(func(rawKey, value string) error {
		key := strings.ToLower(rawKey) // TODO not necessary for plain TextMap
		if strings.HasPrefix(key, p.baggagePrefix) {
			if baggage == nil {
				baggage = make(map[string]string)
			}
			baggage[key[len(p.baggagePrefix):]] = value
		}

		if key != "b3" {
			return nil
		}

		// https://github.com/openzipkin/b3-propagation#single-header
		// b3={TraceId}-{SpanId}-{SamplingState}-{ParentSpanId}, where the last two fields are optional.
		splits := strings.SplitN(value, "-", 4)
		if len(splits) < 2 {
			return fmt.Errorf("invalid b3 value")
		}
		var e error
		if traceID, e = jaeger.TraceIDFromString(splits[0]); e != nil {
			return e
		}

		if spanID, e = jaeger.SpanIDFromString(splits[1]); e != nil {
			return e
		}

		if len(splits) >= 3 {
			if sampled, e = strconv.ParseBool(splits[2]); e != nil {
				return e
			}
		}

		if len(splits) >= 4 {
			if parentID, e = strconv.ParseUint(splits[3], 16, 64); e != nil {
				return e
			}
		}
		return e
	})

	switch {
	case err != nil:
		return jaeger.SpanContext{}, err
	case !traceID.IsValid():
		return jaeger.SpanContext{}, opentracing.ErrSpanContextNotFound
	default:
		return jaeger.NewSpanContext(traceID, spanID, jaeger.SpanID(parentID), sampled, baggage), nil
	}
}
