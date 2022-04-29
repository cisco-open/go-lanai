package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
)

type globalClientProviderFunc func() (sarama.Client, error)
type clusterAdminProviderFunc func() (sarama.ClusterAdmin, error)

type saramaTopicProvisioner struct {
	globalClient globalClientProviderFunc
	adminClient  clusterAdminProviderFunc
}

func (p *saramaTopicProvisioner) topicExists(topic string) (bool, error) {
	gc, e := p.globalClient()
	if e != nil {
		return false, e
	}
	if e := gc.RefreshMetadata(); e != nil {
		return false, translateSaramaBindingError(e, "unable to refresh metadata: %v", e)
	}

	topics, e := gc.Topics()
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

func (p *saramaTopicProvisioner) provisionTopic(topic string, cfg *bindingConfig) error {
	exists, e := p.topicExists(topic)
	if e != nil {
		return e
	}

	if exists {
		return p.tryProvisionPartitions(topic, &cfg.producer.provisioning)
	} else {
		return p.tryCreateTopic(topic, &cfg.producer.provisioning)
	}
}

func (p *saramaTopicProvisioner) tryCreateTopic(topic string, cfg *topicConfig) error {
	if !cfg.autoCreateTopic {
		return NewKafkaError(ErrorCodeIllegalState, fmt.Sprintf(`kafka topic "%s" doesn't exists, and auto-create is disabled`, topic))
	}

	topicDetails := &sarama.TopicDetail{
		NumPartitions:     cfg.partitionCount,
		ReplicationFactor: cfg.replicationFactor,
	}
	ac, e := p.adminClient()
	if e != nil {
		return e
	}
	if e := ac.CreateTopic(topic, topicDetails, false); e != nil {
		return NewKafkaError(ErrorCodeAutoCreateTopicFailed, fmt.Sprintf(`unable to create topic "%s": %v`, topic, e))
	}
	return nil
}

func (p *saramaTopicProvisioner) tryProvisionPartitions(topic string, cfg *topicConfig) error {
	gc, e := p.globalClient()
	if e != nil {
		return e
	}

	parts, e := gc.Partitions(topic)
	if e != nil {
		return translateSaramaBindingError(e, "unable to read partitions config of topic %s: %v", topic, e)
	}

	count := len(parts)
	switch {
	case count >= int(cfg.partitionCount):
		return nil
	case !cfg.autoAddPartitions && cfg.allowLowerPartitions:
		return nil
	case !cfg.autoAddPartitions:
		return NewKafkaError(ErrorCodeAutoAddPartitionsFailed,
			fmt.Sprintf(`topic "%s" has less partitions than required (expected=%d, actual=%d), but auto-add partitions is disabled`,
				topic, cfg.partitionCount, count))
	}

	// we can create partitions
	ac, e := p.adminClient()
	if e != nil {
		return e
	}
	if e := ac.CreatePartitions(topic, cfg.partitionCount, nil, true); e != nil {
		return NewKafkaError(ErrorCodeAutoAddPartitionsFailed, fmt.Sprintf(`unable to add partitions to topic "%s": %v`, topic, e))
	}
	return nil
}