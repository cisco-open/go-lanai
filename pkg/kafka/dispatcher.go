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

package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Shopify/sarama"
	"reflect"
	"strings"
)

// MessageHandlerFunc is message handling function that conform with following signature:
//
//		func (ctx context.Context, [OPTIONAL_INPUT_PARAMS...]) error
//
//	 Where OPTIONAL_INPUT_PARAMS could contain following components (of which order is not important):
//		- PAYLOAD_PARAM 	< AnyPayloadType >: 	message payload, where PayloadType could be any type other than interface, function or chan.
//			                         				If PayloadType is interface{}, raw []byte will be used
//		- HEADERS_PARAM 	< Headers >: 			message headers
//		- METADATA_PARAM 	< *MessageMetadata >: 	message metadata, includes timestamp, keys, partition, etc.
//		- MESSAGE_PARAM 	< *Message >: 			raw message, where Message.Payload would be PayloadType if PAYLOAD_PARAM is also present, or []byte
//
// For Example:
//
//	func Handle(ctx context.Context, payload *MyStruct) error
//	func Handle(ctx context.Context, payload *MyStruct, meta *MessageMetadata) error
//	func Handle(ctx context.Context, payload map[string]interface{}) error
//	func Handle(ctx context.Context, headers Headers, payload *MyStruct) error
//	func Handle(ctx context.Context, payload *MyStruct, raw *Message) error
//	func Handle(ctx context.Context, raw *Message) error
type MessageHandlerFunc interface{}

type MessageFilterFunc func(ctx context.Context, msg *Message) (shouldHandle bool)

var (
	reflectTypeContext  = reflect.TypeOf((*context.Context)(nil)).Elem()
	reflectTypeHeaders  = reflect.TypeOf(Headers{})
	reflectTypeMetadata = reflect.TypeOf(&MessageMetadata{})
	reflectTypeMessage  = reflect.TypeOf(&Message{})
	reflectTypeError    = reflect.TypeOf((*error)(nil)).Elem()
)

type param struct {
	i int
	t reflect.Type
}

func (p param) assign(params []reflect.Value, v reflect.Value) error {
	if p.i >= len(params) || p.t == nil {
		return nil
	}
	if !v.Type().ConvertibleTo(p.t) {
		return ErrorSubTypeIllegalConsumerUsage.WithMessage("failed to prepare parameters for message handler: cannot assign %T to %T", v.String(), p.t.String())
	}
	params[p.i] = v.Convert(p.t)
	return nil
}

type params struct {
	count    int
	payload  param
	headers  param
	metadata param
	message  param
}

type handler struct {
	fn           reflect.Value
	params       params
	filterFunc   MessageFilterFunc
	interceptors []ConsumerHandlerInterceptor
}

type saramaDispatcher struct {
	handlers     []*handler
	interceptors []ConsumerDispatchInterceptor
	msgLogger    MessageLogger
}

func newSaramaDispatcher(cfg *bindingConfig) *saramaDispatcher {
	return &saramaDispatcher{
		handlers:     []*handler{},
		interceptors: cfg.consumer.dispatchInterceptors,
		msgLogger:    cfg.msgLogger,
	}
}

func (d *saramaDispatcher) addHandler(fn MessageHandlerFunc, cfg *consumerConfig, opts []DispatchOptions) error {
	if fn == nil {
		return nil
	}

	// apply options
	f := reflect.ValueOf(fn)
	h := handler{
		fn:           f,
		interceptors: cfg.handlerInterceptors,
	}
	for _, optFn := range opts {
		optFn(&h)
	}

	// parse and validate input params
	t := f.Type()
	for i := t.NumIn() - 1; i >= 0; i-- {
		switch it := t.In(i); {
		case it.AssignableTo(reflectTypeContext):
			if i != 0 {
				return ErrorSubTypeIllegalConsumerUsage.WithMessage("invalid MessageHandlerFunc signature %v, first input param must be context.Context", fn)
			}
		case it.ConvertibleTo(reflectTypeHeaders):
			h.params.headers = param{i, it}
		case it.ConvertibleTo(reflectTypeMetadata):
			h.params.metadata = param{i, it}
		case it.ConvertibleTo(reflectTypeMessage):
			h.params.message = param{i, it}
		case h.params.payload.t == nil && d.isSupportedMessagePayloadType(it):
			h.params.payload = param{i, it}
		default:
			return ErrorSubTypeIllegalConsumerUsage.WithMessage("invalid MessageHandlerFunc signature %v, unknown input parameters at index %v", fn, i)
		}
		h.params.count++
	}

	// parse and validate output params
	for i := t.NumOut() - 1; i >= 0; i-- {
		switch ot := t.Out(i); {
		case ot.ConvertibleTo(reflectTypeError):
			if i != t.NumOut()-1 {
				return ErrorSubTypeIllegalConsumerUsage.WithMessage("invalid MessageHandlerFunc signature %v, last output param must be error", fn)
			}
		default:
			return ErrorSubTypeIllegalConsumerUsage.WithMessage("invalid MessageHandlerFunc signature %v, unknown output parameters at index %v", fn, i)
		}
	}
	d.handlers = append(d.handlers, &h)
	return nil
}

//nolint:contextcheck // context is passed inside msgCtx
func (d saramaDispatcher) dispatch(ctx context.Context, raw *sarama.ConsumerMessage, source interface{}) (err error) {
	defer func() {
		switch e := recover().(type) {
		case error:
			err = ErrorSubTypeConsumerGeneral.WithCause(e, "message dispatcher recovered from panic: %v", e)
		case string:
			err = ErrorSubTypeConsumerGeneral.WithMessage("message dispatcher recovered from panic: %v", e)
		}
	}()

	// parse header
	headers := Headers{}
	for _, rh := range raw.Headers {
		if rh == nil || len(rh.Key) == 0 || len(rh.Value) == 0 {
			continue
		}
		headers[string(rh.Key)] = string(rh.Value)
	}

	// create message context
	msgCtx := &MessageContext{
		Context: ctx,
		Message: Message{
			Headers: headers,
			Payload: raw.Value,
		},
		Source:     source,
		Topic:      raw.Topic,
		RawMessage: raw,
	}

	// invoke interceptors
	for _, interceptor := range d.interceptors {
		msgCtx, err = interceptor.Intercept(msgCtx)
		if err != nil {
			return ErrorSubTypeConsumerGeneral.WithMessage("consumer dispatch interceptor error: %v", err)
		}
	}

	defer func() {
		err = d.finalizeDispatch(msgCtx, err)
	}()

	// log message
	if d.msgLogger != nil {
		d.msgLogger.LogReceivedMessage(msgCtx.Context, raw)
	}

	for _, h := range d.handlers {
		// reset message payload
		msgCtx.Message.Payload = raw.Value

		// filter
		if h.filterFunc != nil {
			if ok := h.filterFunc(msgCtx.Context, &msgCtx.Message); !ok {
				continue
			}
		}

		if err = d.doDispatch(msgCtx, h); err != nil {
			return
		}
	}
	return nil
}

func (d saramaDispatcher) doDispatch(msgCtx *MessageContext, h *handler) (err error) {
	// invoke handler interceptors
	ctx, msg := msgCtx.Context, &msgCtx.Message
	for _, interceptor := range h.interceptors {
		ctx, err = interceptor.BeforeHandling(ctx, msg)
		if err != nil {
			return ErrorSubTypeConsumerGeneral.WithMessage("consumer handler interceptor error: %v", err)
		}
	}

	defer func() {
		for _, interceptor := range h.interceptors {
			ctx, err = interceptor.AfterHandling(ctx, msg, err)
		}
	}()

	// decode payload
	if err = d.decodePayload(ctx, h.params.payload.t, msg); err != nil {
		return
	}

	err = d.invokeHandler(ctx, h, msg, msgCtx)
	return
}

func (d saramaDispatcher) finalizeDispatch(msgCtx *MessageContext, err error) error {
	for _, interceptor := range d.interceptors {
		switch finalizer := interceptor.(type) {
		case ConsumerDispatchFinalizer:
			msgCtx, err = finalizer.Finalize(msgCtx, err)
		}
	}
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ErrorCategoryKafka):
		return err
	default:
		return NewKafkaError(ErrorSubTypeCodeConsumerGeneral, err.Error(), err)
	}
}

/********************
	Helpers
 ********************/

func (d saramaDispatcher) decodePayload(_ context.Context, typ reflect.Type, msg *Message) error {
	if _, ok := msg.Payload.([]byte); !ok || typ == nil {
		return nil
	}
	contentType := msg.Headers[HeaderContentType]
	switch {
	case strings.HasPrefix(contentType, "application/json"):
		ptr, v := d.instantiateByType(typ)
		if e := json.Unmarshal(msg.Payload.([]byte), ptr.Interface()); e != nil {
			return ErrorSubTypeDecoding.WithCause(e, "unable to decode as JSON: %v", e)
		}
		msg.Payload = v.Interface()
	case contentType == MIMETypeText:
		msg.Payload = string(msg.Payload.([]byte))
	case contentType == MIMETypeBinary:
		//  do nothing
	default:
		return ErrorSubTypeDecoding.WithMessage("unsupported MIME type %s", contentType)
	}
	return nil
}

func (d saramaDispatcher) invokeHandler(ctx context.Context, handler *handler, msg *Message, msgCtx *MessageContext) (err error) {
	// prepare input params
	in := make([]reflect.Value, handler.params.count)
	in[0] = reflect.ValueOf(ctx)
	if e := handler.params.payload.assign(in, reflect.ValueOf(msg.Payload)); e != nil {
		return e
	}
	if e := handler.params.headers.assign(in, reflect.ValueOf(msg.Headers)); e != nil {
		return e
	}
	if e := handler.params.message.assign(in, reflect.ValueOf(msg)); e != nil {
		return e
	}

	// message metadata
	if handler.params.metadata.i != 0 {
		var meta *MessageMetadata
		switch raw := msgCtx.RawMessage.(type) {
		case *sarama.ConsumerMessage:
			meta = &MessageMetadata{
				Key:       raw.Key,
				Partition: int(raw.Partition),
				Offset:    int(raw.Offset),
				Timestamp: raw.Timestamp,
			}
		default:
			meta = &MessageMetadata{}
		}

		if e := handler.params.metadata.assign(in, reflect.ValueOf(meta)); e != nil {
			return e
		}
	}

	// invoke
	out := handler.fn.Call(in)

	// post process output
	err, _ = out[0].Interface().(error)
	return
}

// instantiateByType
// "ptr" is the pointer regardless if given type is Ptr or other type
// "value" is actually the value with given type
func (d saramaDispatcher) instantiateByType(t reflect.Type) (ptr reflect.Value, value reflect.Value) {
	switch t.Kind() {
	case reflect.Ptr:
		pp := reflect.New(t)
		p, v := d.instantiateByType(t.Elem())
		pp.Elem().Set(v.Addr())
		return p, pp.Elem()
	default:
		p := reflect.New(t)
		return p, p.Elem()
	}
}

func (d saramaDispatcher) isSupportedMessagePayloadType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr:
		if t.Elem().Kind() == reflect.Ptr {
			return false
		}
		return d.isSupportedMessagePayloadType(t.Elem())
	case reflect.Interface, reflect.Func, reflect.Chan:
		return false
	default:
		return true
	}
}
