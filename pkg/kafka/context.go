package kafka

import (
	"context"
	"github.com/Shopify/sarama"
)

const (
	HeaderContentType = "contentType"
)

const (
	MIMETypeJson   = "application/json;charset=utf-8"
	MIMETypeBinary = "application/octet-stream"
	MIMETypeText   = "text/plain"
)

type Headers map[string]string

type Message struct {
	Headers Headers
	Payload interface{}
}

type Encoder interface {
	// MIMEType returns the MIME type value when using the Encoder
	MIMEType() string
	Encode(v interface{}) ([]byte, error)
}

// MessageContext internal use only, used by interceptors and processors
type MessageContext struct {
	context.Context
	Message
	messageConfig
	Topic string
}

type Binder interface {
	NewProducerWithTopic(topic string, options ...ProducerOptions) (Producer, error)
	ListTopics() []string
}

type BinderLifecycle interface {
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type SaramaBinder interface {
	Binder
	Client() sarama.Client
}

type Producer interface {
	// SendMessage send given message to the TOPIC with pre-configured producer settings
	// supported message types are:
	// 	- *Message
	// 	- Message
	//  - any type of body, the body will be serialized using value encoder from options
	SendMessage(ctx context.Context, message interface{}, options ...MessageOptions) error
}

type ProducerInterceptor interface {
	// Intercept is called before raw message is prepared and send.
	// Implementations can modify fields of MessageContext to manipulate sending behaviour.
	// When error is returned, Producer would cancel operation. Otherwise, a non-nil MessageContext must be returned
	Intercept(msgCtx *MessageContext) (*MessageContext, error)
}
