package kafka

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const (
	ConfigKafkaPrefix               = "kafka"
	ConfigKafkaBindingPrefix        = "kafka.bindings"
	ConfigKafkaDefaultBindingPrefix = "kafka.bindings.default"
)

//goland:noinspection GoNameStartsWithPackageName
type KafkaProperties struct {
	Brokers  utils.CommaSeparatedSlice `json:"brokers"`
	Net      Net                       `json:"net"`
	Metadata Metadata                  `json:"metadata"`
	Binder   BinderProperties          `json:"binder"`
	ClientId string                    `json:"client-id"`
}

type Net struct {
	Sasl SASL `json:"sasl"`
	Tls  TLS  `json:"tls"`
}

type Metadata struct {
	RefreshFrequency utils.Duration `json:"refresh-frequency"`
}

type SASL struct {
	// Whether or not to use SASL authentication when connecting to the broker
	// (defaults to false).
	Enable bool `json:"enabled"`
	// Whether or not to send the Kafka SASL handshake first if enabled
	// (defaults to true). You should only set this to false if you're using
	// a non-Kafka SASL proxy.
	Handshake bool `json:"handshake"`
	//username and password for SASL/PLAIN authentication
	User     string `json:"user"`
	Password string `josn:"password"`
}

type TLS struct {
	Enable bool                       `json:"enabled"`
	Config tlsconfig.SourceProperties `json:"config"`
}

type BinderProperties struct {
	InitialHeartbeat       utils.Duration `json:"init-heartbeat"`
	HeartbeatCurveFactor   float64        `json:"heartbeat-curve-factor"`
	HeartbeatCurveMidpoint float64        `json:"heartbeat-curve-midpoint"`
	WatchdogHeartbeat      utils.Duration `json:"watchdog-heartbeat"`
}

const (
	AckModeModeAll   AckMode = "all"
	AckModeModeLocal AckMode = "local"
	AckModeModeNone  AckMode = "none"
)

type AckMode string

func (m *AckMode) UnmarshalText(data []byte) error {
	switch strings.ToLower(string(data)) {
	case string(AckModeModeAll):
		*m = AckModeModeAll
	case string(AckModeModeLocal):
		*m = AckModeModeLocal
	case string(AckModeModeNone):
		*m = AckModeModeNone
	default:
		*m = AckModeModeNone
	}
	return nil
}

type BindingProperties struct {
	Producer ProducerProperties `json:"producer"`
	Consumer ConsumerProperties `json:"consumer"`
}

type ProducerProperties struct {
	LogLevel     *log.LoggingLevel      `json:"log-level"`
	AckMode      *AckMode               `json:"ack-mode"`
	AckTimeout   *utils.Duration        `json:"ack-timeout"`
	MaxRetry     *int                   `json:"max-retry"`
	Backoff      *utils.Duration        `json:"backoff-interval"`
	Provisioning ProvisioningProperties `json:"provisioning"`
}

type ConsumerProperties struct {
	LogLevel *log.LoggingLevel       `json:"log-level"`
	Backoff  *utils.Duration         `json:"backoff-interval"`
	Group    ConsumerGroupProperties `json:"group"`
}

type ProvisioningProperties struct {
	// AutoCreateTopic when topic doesn't exist, whether attempt to create one
	AutoCreateTopic *bool `json:"auto-create-topic"`

	// AutoAddPartitions when actual partition counts is less than partitionCount, whether attempt to add more partitions
	AutoAddPartitions *bool `json:"auto-add-partitions"`

	// allowLowerPartitions when actual partition counts is less than partitionCount but autoAddPartitions is false,
	// whether return an error
	AllowLowerPartitions *bool `json:"allow-lower-partitions"`

	// PartitionCount number of partitions of given topic
	PartitionCount *int32 `json:"partition-count"`

	// ReplicationFactor number of replicas per partition when creating topic
	ReplicationFactor *int16 `json:"replication-factor"`
}

type ConsumerGroupProperties struct {
	JoinTimeout *utils.Duration `json:"join-timeout"`
	MaxRetry    *int            `json:"max-retry"`
	Backoff     *utils.Duration `json:"backoff-interval"`
}

func BindKafkaProperties(ctx *bootstrap.ApplicationContext) KafkaProperties {
	props := KafkaProperties{
		Net: Net{
			Sasl: SASL{
				Enable:    false,
				Handshake: true,
			},
			Tls: TLS{
				Enable: false,
			},
		},
		Metadata: Metadata{
			RefreshFrequency: utils.Duration(5 * time.Minute),
		},
		Binder: BinderProperties{
			InitialHeartbeat:       utils.Duration(5 * time.Second),
			WatchdogHeartbeat:      utils.Duration(120 * time.Second),
			HeartbeatCurveFactor:   0.5,
			HeartbeatCurveMidpoint: 10, // recommend > 5
		},
		ClientId: ctx.Name(),
	}
	if err := ctx.Config().Bind(&props, ConfigKafkaPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind kafka properties"))
	}
	return props
}
