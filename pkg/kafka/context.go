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
	"github.com/IBM/sarama"
	"time"
)

const (
	HeaderContentType = "contentType"
)

const (
	MIMETypeJson   = "application/json;charset=utf-8"
	MIMETypeBinary = "application/octet-stream"
	MIMETypeText   = "text/plain"
)

/************************
	General
 ************************/

type Headers map[string]string

type Message struct {
	Headers Headers
	Payload interface{}
}

type MessageMetadata struct {
	Key       []byte
	Partition int
	Offset    int
	Timestamp time.Time
}

type Encoder interface {
	// MIMEType returns the MIME type value when using the Encoder
	MIMEType() string
	Encode(v interface{}) ([]byte, error)
}

// MessageContext internal use only, used by Interceptors and processors
type MessageContext struct {
	context.Context
	messageConfig
	Source     interface{}
	Topic      string
	Message    Message
	RawMessage interface{}
}

/************************
	Binding
 ************************/

type ProducerOptions func(cfg *bindingConfig)

type ConsumerOptions func(cfg *bindingConfig)

type Binder interface {
	// Produce create a message Producer to given topic, if not existed already. This would also auto-create topics if configured correctly
	// The returned Producer may also implement BindingLifecycle, in which case BindingLifecycle.Close can be used to manually release resources
	Produce(topic string, options ...ProducerOptions) (Producer, error)

	// Subscribe create a Subscriber of given topic, if not existed already.
	// The returned Subscriber may also implement BindingLifecycle, in which case BindingLifecycle.Close can be used to manually release resources
	Subscribe(topic string, options ...ConsumerOptions) (Subscriber, error)

	// Consume create a GroupConsumer of given topic and group, if not existed already.
	// The returned GroupConsumer may also implement BindingLifecycle, in which case BindingLifecycle.Close can be used to manually release resources
	Consume(topic string, group string, options ...ConsumerOptions) (GroupConsumer, error)

	// ListTopics list of topics of all managed bindings
	ListTopics() []string
}

type BinderLifecycle interface {
	// Initialize should be called only once, before Shutdown is executed.
	Initialize(ctx context.Context) error
	// Start should be called only once, before Shutdown is executed.
	// If the provided context is cancellable, the lifecycle is shutdown automatically when it happens
	Start(ctx context.Context) error
	// Shutdown can be called multiple times, but would do nothing if called more than once.
	// Important:
	// - Once called, Calling Initialize or Start would cause unexpected error or resource leak
	// - Once called, any Producer, Subscriber or GroupConsumer issued by this Binder are closed automatically.
	//	 They should be discarded
	Shutdown(ctx context.Context) error
	// Done channel used for monitoring lifecycle. Channel is closed during shutdown
	Done() <-chan struct{}
}

type SaramaBinder interface {
	Binder
	Client() sarama.Client
}

// BindingLifecycle is the interface that controlling lifecycles of any Producer, Subscriber and GroupConsumer
type BindingLifecycle interface {
	// Start initialize any connection and internal run loops.
	// Must be called after any configuration and before Producer, Subscriber or GroupConsumer is used
	Start(ctx context.Context) error

	// Close must be called to release any resource. Once called, regardless the return, this instance must be discarded
	Close() error

	// Closed returns if Close has been called
	Closed() bool
}

/************************
	Producing
 ************************/

type Producer interface {
	// Topic returns the Topic name
	Topic() string

	// Send publish given message to the TOPIC with pre-configured producer settings
	// supported message types are:
	// 	- *Message
	// 	- Message
	//  - any type of body, the body will be serialized using value encoder from options
	Send(ctx context.Context, message interface{}, options ...MessageOptions) error

	// ReadyCh is used to monitor producer status. The channel is closed once the Producer is ready to send messages.
	ReadyCh() <-chan struct{}
}

type ProducerMessageInterceptor interface {
	// Intercept is called before raw message is prepared and send.
	// Implementations can modify fields of MessageContext to manipulate sending behaviour.
	// When error is returned, Producer would cancel operation. Otherwise, a non-nil MessageContext must be returned
	Intercept(msgCtx *MessageContext) (*MessageContext, error)
}

// ProducerMessageFinalizer is the interface for other package to finalize message sending process.
// When any ProducerMessageInterceptor also implements ProducerMessageFinalizer, the Finalize function will be invoked
// after message delivery is confirmed
type ProducerMessageFinalizer interface {
	// Finalize is called after message delivery is confirmed.
	// "confirmed" status depends on Ack mode of the message.
	// e.g.
	// 	- if the message uses RequireNoAck, Finalize is called right after sending the message
	// 	- if the message uses RequireAllAck, Finalize is called when Ack is received from all replicas
	//
	// Finalize may also be invoked in different goroutine if delivery mode is "sync"
	//
	// Note: the *MessageContext will be discarded after all finalizers finished processing.
	// 		 So modifying given message context would only affect subsequent ProducerMessageFinalizer.Finalize on same message
	// Note 2: Finalize may also choose to handle given err and returns nil error.
	//		   In such case, subsequent ProducerMessageFinalizer.Finalize on same message would be invoked as if there was no error
	Finalize(msgCtx *MessageContext, partition int32, offset int64, err error) (*MessageContext, error)
}

/************************
	Consuming
 ************************/

// Subscriber provides Pub-Sub workflow
type Subscriber interface {
	// Topic returns the Topic name
	Topic() string

	// Partitions returns subscribed partitions
	Partitions() []int32

	// AddHandler register a message handler function that would process received messages.
	// Note: A Subscriber without a registered handler simply ignore all received messages.
	// 		 If AddHandler is called after Start, it may miss some messages
	AddHandler(handlerFunc MessageHandlerFunc, opts ...DispatchOptions) error
}

// GroupConsumer provides consumer group workflow
type GroupConsumer interface {
	// Topic returns the Topic name
	Topic() string

	// Group returns the group name
	Group() string

	// AddHandler register a message handler function that would process received messages.
	// Note: A GroupConsumer without a registered handler simply ignore all received messages.
	// 		 If AddHandler is called after Start, it may miss some messages
	AddHandler(handlerFunc MessageHandlerFunc, opts ...DispatchOptions) error
}

type ConsumerDispatchInterceptor interface {
	// Intercept is called before message is decoded and consumed by registered handlers.
	// Implementations can modify fields of MessageContext to manipulate behaviour.
	// When error is returned, dispatcher would cancel operation. Otherwise, a non-nil MessageContext must be returned
	Intercept(msgCtx *MessageContext) (*MessageContext, error)
}

// ConsumerDispatchFinalizer is the interface for other package to finalize message processing.
// When any ConsumerDispatchInterceptor also implements ConsumerDispatchFinalizer, the Finalize function will be invoked
// after message processing finished
type ConsumerDispatchFinalizer interface {
	// Finalize is called after message processing is finished by handlers.
	// Note: the *MessageContext will be discarded after all finalizers finished processing.
	// 		 So modifying given message context would only affect subsequent ConsumerDispatchFinalizer.Finalize on same message
	// Note 2: Finalize may also choose to handle given err and returns nil error.
	//		   In such case, subsequent ConsumerDispatchFinalizer.Finalize on same message would be invoked as if there was no error
	Finalize(msgCtx *MessageContext, err error) (*MessageContext, error)
}

type ConsumerHandlerInterceptor interface {
	// BeforeHandling is called after message is decoded and before handled by each registered handlers.
	// Implementations can modify fields of MessageContext to manipulate behaviour.
	// When error is returned, dispatcher would cancel operation. Otherwise, a non-nil MessageContext must be returned
	BeforeHandling(ctx context.Context, msg *Message) (context.Context, error)

	// AfterHandling is called after each registered handlers handles the message
	AfterHandling(ctx context.Context, msg *Message, err error) (context.Context, error)
}

/************************
	Internals
 ************************/

type bindingConfig struct {
	name       string
	properties BindingProperties
	sarama     sarama.Config
	producer   producerConfig
	consumer   consumerConfig
	msgLogger  MessageLogger
}

type producerConfig struct {
	keyEncoder   Encoder
	interceptors []ProducerMessageInterceptor
	provisioning topicConfig
}

type consumerConfig struct {
	dispatchInterceptors []ConsumerDispatchInterceptor
	handlerInterceptors  []ConsumerHandlerInterceptor
	msgLogger            MessageLogger
}

type topicConfig struct {
	// autoCreateTopic when topic doesn't exist, whether attempt to create one
	autoCreateTopic bool

	// autoAddPartitions when actual partition counts is less than partitionCount, whether attempt to add more partitions
	autoAddPartitions bool

	// allowLowerPartitions when actual partition counts is less than partitionCount but autoAddPartitions is false,
	// whether return an error
	allowLowerPartitions bool

	// partitionCount number of partitions of given topic
	partitionCount int32

	// replicationFactor number of replicas per partition when creating topic
	replicationFactor int16
}
