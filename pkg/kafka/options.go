package kafka

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"time"
)

/********************
 Options for producer
 ********************/
type producerConfig struct {
	*sarama.Config
}

func defaultProducerConfig(properties KafkaProperties) *producerConfig {
	c := &producerConfig{
		Config: sarama.NewConfig(),
	}

	if properties.Net.Sasl.Enable {
		c.Net.SASL.Enable = properties.Net.Sasl.Enable
		c.Net.SASL.Handshake = properties.Net.Sasl.Handshake
		c.Net.SASL.User = properties.Net.Sasl.User
		c.Net.SASL.Password = properties.Net.Sasl.Password
	}

	return c
}

type ProducerOptions func(*producerConfig)

// RequireAllAck waits for all in-sync replicas to commit before responding.
// The minimum number of in-sync replicas is configured on the broker via
// the `min.insync.replicas` configuration key.
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

/********************
 Options for message
********************/
type deliveryMode int

const (
	sync deliveryMode = iota
)

type messageConfig struct {
	valueEncoder func(v interface{})([]byte, error)
	key          interface{}
	keyEncoder   func(v interface{})([]byte, error)
	mode         deliveryMode
}

func defaultMessageConfig() *messageConfig{
	return &messageConfig{
		valueEncoder: json.Marshal,
		keyEncoder: json.Marshal,
		mode: sync,
	}
}

type MessageOptions func(*messageConfig)

func WithKey(key interface{}) MessageOptions {
	return func(config *messageConfig) {
		config.key = key
	}
}