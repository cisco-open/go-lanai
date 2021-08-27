package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
)

type saramaTopicProvisioner struct {
	globalClient   sarama.Client
	adminClient    sarama.ClusterAdmin
}

func (p *saramaTopicProvisioner) topicExists(topic string) (bool, error) {
	topics, e := p.globalClient.Topics()
	if e != nil {
		return false, translateSaramaBindingError(e, "unable to read topics: %v", e)
	}

	for _, t := range topics {
		if t == topic {
			return true, nil
		}
	}
	return false, nil
}

func (p *saramaTopicProvisioner) provisionTopic(topic string, cfg *producerConfig) error {
	exists, e := p.topicExists(topic)
	if e != nil {
		return e
	}

	if exists {
		return p.tryProvisionPartitions(topic, cfg)
	} else {
		return p.tryCreateTopic(topic, cfg)
	}
}

func (p *saramaTopicProvisioner) tryCreateTopic(topic string, cfg *producerConfig) error {
	if !cfg.provisioning.autoCreateTopic {
		return NewKafkaError(ErrorCodeIllegalState, fmt.Sprintf(`kafka topic "%s" doesn't exists, and auto-create is disabled`, topic))
	}

	topicDetails := &sarama.TopicDetail{
		NumPartitions:     cfg.provisioning.partitionCount,
		ReplicationFactor: cfg.provisioning.replicationFactor,
	}
	if e := p.adminClient.CreateTopic(topic, topicDetails, false); e != nil {
		return NewKafkaError(ErrorCodeAutoCreateTopicFailed, fmt.Sprintf(`unable to create topic "%s": %v`, topic, e))
	}
	return nil
}

func (p *saramaTopicProvisioner) tryProvisionPartitions(topic string, cfg *producerConfig) error {
	parts, e := p.globalClient.Partitions(topic)
	if e != nil {
		return translateSaramaBindingError(e, "unable to read partitions config of topic %s: %v", topic, e)
	}

	count := len(parts)
	switch {
	case count >= int(cfg.provisioning.partitionCount):
		return nil
	case !cfg.provisioning.autoAddPartitions && cfg.provisioning.allowLowerPartitions:
		return nil
	case !cfg.provisioning.autoAddPartitions:
		return NewKafkaError(ErrorCodeAutoAddPartitionsFailed,
			fmt.Sprintf(`topic "%s" has less partitions than required (expected=%d, actual=%d), but auto-add partitions is disabled`,
				topic, cfg.provisioning.partitionCount, count))
	}

	// we can create partitions
	if e := p.adminClient.CreatePartitions(topic, cfg.provisioning.partitionCount, nil, true); e != nil {
		return NewKafkaError(ErrorCodeAutoAddPartitionsFailed, fmt.Sprintf(`unable to add partitions to topic "%s": %v`, topic, e))
	}
	return nil
}