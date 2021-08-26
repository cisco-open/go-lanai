package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/google/uuid"
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

type producerConfig struct {
	*sarama.Config
	keyEncoder     Encoder
	partitionCount int32
	replicationFactor int16
	interceptors   []ProducerInterceptor
}

func defaultProducerConfig(saramaCfg *sarama.Config) *producerConfig {
	return &producerConfig{
		Config:         saramaCfg,
		keyEncoder:     binaryEncoder{},
		partitionCount: 1,
		replicationFactor: 0,
		interceptors:   []ProducerInterceptor{},
	}
}

type ProducerOptions func(*producerConfig)

// WithKeyEncoder configures Producer with given encoder for serializing message key
func WithKeyEncoder(enc Encoder) ProducerOptions {
	return func(config *producerConfig) {
		config.keyEncoder = enc
	}
}

// WithPartitions TODO
func WithPartitions(partitionCount int, replicationFactor int) ProducerOptions {
	return func(config *producerConfig) {
		config.partitionCount = int32(partitionCount)
		config.replicationFactor = int16(replicationFactor)
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
	*sarama.Config
}

type ConsumerOptions func(*consumerConfig)

func defaultConsumerConfig(saramaCfg *sarama.Config) *consumerConfig {
	return &consumerConfig{
		Config: saramaCfg,
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
		Key:          uuid.New(),
		Mode:         modeSync,
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

//// WithTypeOf specify message payload type of MessageHandlerFunc
//func WithTypeOf(i interface{}) DispatchOptions {
//	return func(h *handler) {
//		h.typ = reflect.TypeOf(i)
//	}
//}