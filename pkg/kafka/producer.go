package kafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
)

type SaramaProducer struct {
	topic        string
	syncProducer sarama.SyncProducer
}

func NewSaramaProducer(topic string, addrs []string, config *sarama.Config) (*SaramaProducer, error) {
	s, err := newSaramaProducer(topic, addrs, config)

	if err != nil {
		return nil, err
	}
	return s, nil
}

func newSaramaProducer(topic string, addrs []string, config *sarama.Config) (*SaramaProducer, error) {
	c := *config //make a copy so that we don't change the original config
	//sync producer must have these two properties set to true
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true

	internal, err := sarama.NewSyncProducer(addrs, &c)

	if err != nil {
		return nil, err
	}

	p := &SaramaProducer{
		topic:        topic,
		syncProducer: internal,

	}
	return p, nil
}

func (s *SaramaProducer) SendMessage(ctx context.Context, message interface{}, options ...MessageOptions) error {
	config := defaultMessageConfig()
	for _, optionFunc := range options {
		optionFunc(config)
	}

	encodedValue, err := config.valueEncoder(message)
	if err != nil {
		return err
	}

	saramaMessage := &sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.ByteEncoder(encodedValue),
	}

	if config.key != nil {
		encodedKey, err := config.keyEncoder(config.key)
		if err != nil {
			return err
		}
		saramaMessage.Key = sarama.ByteEncoder(encodedKey)
	}

	if config.mode == sync {
		_, _, err = s.syncProducer.SendMessage(saramaMessage)
		return err
	} else {
		return errors.New(fmt.Sprintf("%v mode is not supported", config.mode))
	}
}

func (s *SaramaProducer) Close() error {
	return s.syncProducer.Close()
}