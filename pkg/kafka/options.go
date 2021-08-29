package kafka

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"github.com/Shopify/sarama"
	"time"
)

func defaultSaramaConfig(properties *KafkaProperties) (c *sarama.Config) {
	c = sarama.NewConfig()
	c.Version = sarama.V2_0_0_0

	if properties.Net.Sasl.Enable {
		c.Net.SASL.Enable = properties.Net.Sasl.Enable
		c.Net.SASL.Handshake = properties.Net.Sasl.Handshake
		c.Net.SASL.User = properties.Net.Sasl.User
		c.Net.SASL.Password = properties.Net.Sasl.Password
	}
	return
}

/***********************
  Options for producer
************************/

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

type producerConfig struct {
	sarama.Config
	keyEncoder   Encoder
	interceptors []ProducerMessageInterceptor
	msgLogger    MessageLogger
	provisioning topicConfig
}

type ProducerOptions func(*producerConfig)

// WithProducerProperties apply options configured via ProducerProperties
func WithProducerProperties(p *ProducerProperties) ProducerOptions {
	return func(cfg *producerConfig) {
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
			WithProducerLogLevel(*p.LogLevel)(cfg)
		}
		utils.MustSetIfNotNil(&cfg.Config.Producer.Timeout, p.AckTimeout)
		utils.MustSetIfNotNil(&cfg.Config.Producer.Retry.Max, p.MaxRetry)
		utils.MustSetIfNotNil(&cfg.Config.Producer.Retry.Backoff, p.Backoff)
		utils.MustSetIfNotNil(&cfg.provisioning.autoCreateTopic, p.Provisioning.AutoCreateTopic)
		utils.MustSetIfNotNil(&cfg.provisioning.autoAddPartitions, p.Provisioning.AutoAddPartitions)
		utils.MustSetIfNotNil(&cfg.provisioning.allowLowerPartitions, p.Provisioning.AllowLowerPartitions)
		utils.MustSetIfNotNil(&cfg.provisioning.partitionCount, p.Provisioning.PartitionCount)
		utils.MustSetIfNotNil(&cfg.provisioning.replicationFactor, p.Provisioning.ReplicationFactor)
	}
}

// WithKeyEncoder configures Producer with given encoder for serializing message key
func WithKeyEncoder(enc Encoder) ProducerOptions {
	return func(config *producerConfig) {
		config.keyEncoder = enc
	}
}

// WithPartitions configure Producer's topic provisioning, by specifying min partition required
// and their replica number (min.insync.replicas) in case topics are auto-created
func WithPartitions(partitionCount int, replicationFactor int) ProducerOptions {
	return func(config *producerConfig) {
		if partitionCount < 1 {
			partitionCount = 1
		}
		if replicationFactor < 1 {
			replicationFactor = 1
		}
		config.provisioning.partitionCount = int32(partitionCount)
		config.provisioning.replicationFactor = int16(replicationFactor)
	}
}

// WithProducerLogLevel specify log level of internal message logger
func WithProducerLogLevel(level log.LoggingLevel) ProducerOptions {
	return func(config *producerConfig) {
		config.msgLogger = config.msgLogger.WithLevel(level)
	}
}

// RequireAllAck waits for all in-sync replicas to commit before responding.
// The minimum number of in-sync replicas is configured on the broker via
// the `min.insync.replicas` configuration Key.
func RequireAllAck() ProducerOptions {
	return func(config *producerConfig) {
		config.Producer.RequiredAcks = sarama.WaitForAll
	}
}

// RequireLocalAck waits for only the local commit to succeed before responding.
func RequireLocalAck() ProducerOptions {
	return func(config *producerConfig) {
		config.Producer.RequiredAcks = sarama.WaitForLocal
	}
}

// RequireNoAck doesn't send any response, the TCP ACK is all you get.
func RequireNoAck() ProducerOptions {
	return func(config *producerConfig) {
		config.Producer.RequiredAcks = sarama.NoResponse
	}
}

func AckTimeout(timeout time.Duration) ProducerOptions {
	return func(config *producerConfig) {
		config.Producer.Timeout = timeout
	}
}

/***********************
  Options for consumer
************************/

type consumerConfig struct {
	sarama.Config
	dispatchInterceptors []ConsumerDispatchInterceptor
	handlerInterceptors  []ConsumerHandlerInterceptor
	msgLogger            MessageLogger
}

type ConsumerOptions func(*consumerConfig)

// WithConsumerProperties apply options configured via ConsumerProperties
func WithConsumerProperties(p *ConsumerProperties) ConsumerOptions {
	return func(cfg *consumerConfig) {
		if p.LogLevel != nil {
			WithConsumerLogLevel(*p.LogLevel)(cfg)
		}
		utils.MustSetIfNotNil(&cfg.Consumer.Retry.Backoff, p.Backoff)
		utils.MustSetIfNotNil(&cfg.Consumer.Group.Rebalance.Timeout, p.Group.JoinTimeout)
		utils.MustSetIfNotNil(&cfg.Consumer.Group.Rebalance.Retry.Max, p.Group.MaxRetry)
		utils.MustSetIfNotNil(&cfg.Consumer.Group.Rebalance.Retry.Backoff, p.Group.Backoff)
	}
}

// WithConsumerLogLevel specify log level of internal message logger
func WithConsumerLogLevel(level log.LoggingLevel) ConsumerOptions {
	return func(config *consumerConfig) {
		config.msgLogger = config.msgLogger.WithLevel(level)
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
// Supported values depends on the WithKeyEncoder option on the Producer.
// Default encoder support following types:
// 	- uuid.UUID
// 	- string
// 	- []byte
// 	- encoding.BinaryMarshaler
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
		return func(h *handler) {}
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
