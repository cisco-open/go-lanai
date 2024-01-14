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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"github.com/IBM/sarama"
	"time"
)

func defaultSaramaConfig(_ context.Context, properties *KafkaProperties) (c *sarama.Config, err error) {
	c = sarama.NewConfig()
	c.Version = sarama.V2_0_0_0

	c.ClientID = properties.ClientId
	c.Metadata.RefreshFrequency = time.Duration(properties.Metadata.RefreshFrequency)

	if properties.Net.Sasl.Enable {
		c.Net.SASL.Enable = properties.Net.Sasl.Enable
		c.Net.SASL.Handshake = properties.Net.Sasl.Handshake
		c.Net.SASL.User = properties.Net.Sasl.User
		c.Net.SASL.Password = properties.Net.Sasl.Password
	}
	return
}

/*************************************
  Options for Producer and Consumer
**************************************/

// Note: the return type here have to be unnamed func for compiler to accept as both ProducerOptions and ConsumerOptions
// 		 See https://golang.org/ref/spec#Type_identity

// BindingName is a ProducerOptions or ConsumerOptions that specify the name of the binding.
// This name is used to read BindingProperties from bootstrap.ApplicationConfig
// If not specified, lower case of topic name is used.
// Regardless if name is specified or if corresponding BindingProperties is found,
// any ProducerOptions or ConsumerOptions used at compile time still apply.
// The overriding order is as follows:
//
//	BindingProperties with matching name >
//	BindingProperties with name "default" >
//	ProducerOptions or ConsumerOptions >
//	prepared defaults during initialization
func BindingName(name string) func(cfg *bindingConfig) {
	return func(config *bindingConfig) {
		if name != "" {
			config.name = name
		}
	}
}

// LogLevel is a ProducerOptions or ConsumerOptions that specify log level of Producer, Subscriber or Consumer
func LogLevel(level log.LoggingLevel) func(cfg *bindingConfig) {
	return func(config *bindingConfig) {
		config.msgLogger = config.msgLogger.WithLevel(level)
	}
}

/***********************
  Options for producer
************************/

// WithProducerProperties apply options configured via ProducerProperties
func WithProducerProperties(p *ProducerProperties) ProducerOptions {
	return func(cfg *bindingConfig) {
		if p.AckMode != nil {
			switch *p.AckMode {
			case AckModeModeAll:
				RequireAllAck()(cfg)
			case AckModeModeLocal:
				RequireLocalAck()(cfg)
			case AckModeModeNone:
				RequireNoAck()(cfg)
			}
		}

		if p.LogLevel != nil {
			LogLevel(*p.LogLevel)(cfg)
		}
		utils.MustSetIfNotNil(&cfg.sarama.Producer.Timeout, p.AckTimeout)
		utils.MustSetIfNotNil(&cfg.sarama.Producer.Retry.Max, p.MaxRetry)
		utils.MustSetIfNotNil(&cfg.sarama.Producer.Retry.Backoff, p.Backoff)
		utils.MustSetIfNotNil(&cfg.producer.provisioning.autoCreateTopic, p.Provisioning.AutoCreateTopic)
		utils.MustSetIfNotNil(&cfg.producer.provisioning.autoAddPartitions, p.Provisioning.AutoAddPartitions)
		utils.MustSetIfNotNil(&cfg.producer.provisioning.allowLowerPartitions, p.Provisioning.AllowLowerPartitions)
		utils.MustSetIfNotNil(&cfg.producer.provisioning.partitionCount, p.Provisioning.PartitionCount)
		utils.MustSetIfNotNil(&cfg.producer.provisioning.replicationFactor, p.Provisioning.ReplicationFactor)
	}
}

// KeyEncoder configures Producer with given encoder for serializing message key
func KeyEncoder(enc Encoder) ProducerOptions {
	return func(config *bindingConfig) {
		config.producer.keyEncoder = enc
	}
}

// Partitions configure Producer's topic provisioning, by specifying min partition required
// and their replica number (min.insync.replicas) in case topics are auto-created
func Partitions(partitionCount int, replicationFactor int) ProducerOptions {
	return func(config *bindingConfig) {
		if partitionCount < 1 {
			partitionCount = 1
		}
		if replicationFactor < 1 {
			replicationFactor = 1
		}
		config.producer.provisioning.partitionCount = int32(partitionCount)
		config.producer.provisioning.replicationFactor = int16(replicationFactor)
	}
}

// RequireAllAck waits for all in-sync replicas to commit before responding.
// The minimum number of in-sync replicas is configured on the broker via
// the `min.insync.replicas` configuration Key.
func RequireAllAck() ProducerOptions {
	return func(config *bindingConfig) {
		config.sarama.Producer.RequiredAcks = sarama.WaitForAll
	}
}

// RequireLocalAck waits for only the local commit to succeed before responding.
func RequireLocalAck() ProducerOptions {
	return func(config *bindingConfig) {
		config.sarama.Producer.RequiredAcks = sarama.WaitForLocal
	}
}

// RequireNoAck doesn't send any response, the TCP ACK is all you get.
func RequireNoAck() ProducerOptions {
	return func(config *bindingConfig) {
		config.sarama.Producer.RequiredAcks = sarama.NoResponse
	}
}

func AckTimeout(timeout time.Duration) ProducerOptions {
	return func(config *bindingConfig) {
		config.sarama.Producer.Timeout = timeout
	}
}

/***********************
  Options for consumer
************************/

// WithConsumerProperties apply options configured via ConsumerProperties
func WithConsumerProperties(p *ConsumerProperties) ConsumerOptions {
	return func(cfg *bindingConfig) {
		if p.LogLevel != nil {
			LogLevel(*p.LogLevel)(cfg)
		}
		utils.MustSetIfNotNil(&cfg.sarama.Consumer.Retry.Backoff, p.Backoff)
		utils.MustSetIfNotNil(&cfg.sarama.Consumer.Group.Rebalance.Timeout, p.Group.JoinTimeout)
		utils.MustSetIfNotNil(&cfg.sarama.Consumer.Group.Rebalance.Retry.Max, p.Group.MaxRetry)
		utils.MustSetIfNotNil(&cfg.sarama.Consumer.Group.Rebalance.Retry.Backoff, p.Group.Backoff)
	}
}

/**********************
  Options for message
***********************/

type deliveryMode int

const (
	modeSync deliveryMode = iota
)

type messageConfig struct {
	ValueEncoder Encoder
	Key          interface{}
	Mode         deliveryMode
}

func defaultMessageConfig() messageConfig {
	return messageConfig{
		ValueEncoder: jsonEncoder{},
		//Key:          uuid.New(),
		Mode: modeSync,
	}
}

type MessageOptions func(config *messageConfig)

// WithKey specify key used for the message. The key is typically used for partitioning.
// Supported values depends on the KeyEncoder option on the Producer.
// Default encoder support following types:
//   - uuid.UUID
//   - string
//   - []byte
//   - encoding.BinaryMarshaler
func WithKey(key interface{}) MessageOptions {
	return func(config *messageConfig) {
		config.Key = key
	}
}

// WithEncoder specify how message payload is encoded.
// Default is "application/json;application/json;charset=utf-8"
func WithEncoder(valueEncoder Encoder) MessageOptions {
	return func(config *messageConfig) {
		config.ValueEncoder = valueEncoder
	}
}

/*************************
  Options for dispatcher
**************************/

type DispatchOptions func(h *handler)

// FilterOnHeader returns a DispatchOptions specifying that
// the handler should be invoked when certain message header exists and matches the provided matcher
func FilterOnHeader(header string, matcher matcher.StringMatcher) DispatchOptions {
	if matcher == nil {
		return noop()
	}

	return func(h *handler) {
		h.filterFunc = func(ctx context.Context, msg *Message) (shouldHandle bool) {
			if msg.Headers == nil {
				return false
			}
			v, ok := msg.Headers[header]
			if !ok {
				return false
			}
			if matched, e := matcher.MatchesWithContext(ctx, v); e != nil || !matched {
				return false
			}
			return true
		}
	}
}

func noop() func(h *handler) {
	return func(_ *handler) {
		// noop
	}
}
