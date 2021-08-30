package kafka

import (
	. "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	errorutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
)

const (
	// Reserved kafka reserved error range
	Reserved = 0x1a << errorutils.ReservedOffset
)

// All "Type" values are used as mask
const (
	_                    = iota
	ErrorTypeCodeBinding = Reserved + iota<<errorutils.ErrorTypeOffset
	ErrorTypeCodeProducer
	ErrorTypeCodeConsumer
)

// All "SubType" values are used as mask
// sub-types of ErrorTypeCodeBinding
const (
	_                               = iota
	ErrorSubTypeCodeBindingInternal = ErrorTypeCodeBinding + iota<<errorutils.ErrorSubTypeOffset
	ErrorSubTypeCodeConnectivity
	ErrorSubTypeCodeProvisioning
)

// ErrorSubTypeCodeBindingInternal
const (
	_                        = iota
	ErrorCodeBindingInternal = ErrorSubTypeCodeBindingInternal + iota
)

// ErrorSubTypeCodeConnectivity
const (
	_                           = iota
	ErrorCodeBrokerNotReachable = ErrorSubTypeCodeConnectivity + iota
)

// ErrorSubTypeCodeProvisioning
const (
	_                     = iota
	ErrorCodeIllegalState = ErrorSubTypeCodeProvisioning + iota
	ErrorCodeProducerExists
	ErrorCodeConsumerExists
	ErrorCodeAutoCreateTopicFailed
	ErrorCodeAutoAddPartitionsFailed
)

// All "SubType" values are used as mask
// sub-types of ErrorTypeProducer
const (
	_                               = iota
	ErrorSubTypeCodeProducerGeneral = ErrorTypeCodeProducer + iota<<errorutils.ErrorSubTypeOffset
	ErrorSubTypeCodeIllegalProducerUsage
	ErrorSubTypeCodeEncoding
)

// All "SubType" values are used as mask
// sub-types of ErrorTypeConsumer
const (
	_                               = iota
	ErrorSubTypeCodeConsumerGeneral = ErrorTypeCodeConsumer + iota<<errorutils.ErrorSubTypeOffset
	ErrorSubTypeCodeIllegalConsumerUsage
	ErrorSubTypeCodeDecoding
)

// ErrorTypes, can be used in errors.Is
var (
	ErrorCategoryKafka = NewErrorCategory(Reserved, errors.New("error type: kafka"))
	ErrorTypeBinding   = NewErrorType(ErrorTypeCodeBinding, errors.New("error type: authentication"))
	ErrorTypeProducer  = NewErrorType(ErrorTypeCodeProducer, errors.New("error type: access control"))
	ErrorTypeConsumer  = NewErrorType(ErrorTypeCodeConsumer, errors.New("error type: internal"))

	ErrorSubTypeBindingInternal = NewErrorSubType(ErrorSubTypeCodeBindingInternal, errors.New("error sub-type: internal"))
	ErrorSubTypeConnectivity    = NewErrorSubType(ErrorSubTypeCodeConnectivity, errors.New("error sub-type: internal"))
	ErrorSubTypeProvisioning    = NewErrorSubType(ErrorSubTypeCodeProvisioning, errors.New("error sub-type: internal"))

	ErrorSubTypeProducerGeneral      = NewErrorSubType(ErrorSubTypeCodeProducerGeneral, errors.New("error sub-type: csrf"))
	ErrorSubTypeIllegalProducerUsage = NewErrorSubType(ErrorSubTypeCodeIllegalProducerUsage, errors.New("error sub-type: csrf"))
	ErrorSubTypeEncoding             = NewErrorSubType(ErrorSubTypeCodeEncoding, errors.New("error sub-type: csrf"))
	ErrorSubTypeConsumerGeneral      = NewErrorSubType(ErrorSubTypeCodeConsumerGeneral, errors.New("error sub-type: csrf"))
	ErrorSubTypeIllegalConsumerUsage = NewErrorSubType(ErrorSubTypeCodeIllegalConsumerUsage, errors.New("error sub-type: csrf"))
	ErrorSubTypeDecoding             = NewErrorSubType(ErrorSubTypeCodeDecoding, errors.New("error sub-type: csrf"))
)

func init() {
	errorutils.Reserve(ErrorCategoryKafka)
}

/************************
	Constructors
*************************/

func NewKafkaError(code int64, text string, causes ...interface{}) error {
	return NewCodedError(code, errors.New(text), causes...)
}

func translateSaramaBindingError(cause error, msg string, args ...interface{}) error {
	if errors.Is(cause, ErrorCategoryKafka) {
		return cause
	}
	switch cause {
	case sarama.ErrOutOfBrokers:
		return NewKafkaError(ErrorCodeBrokerNotReachable, fmt.Sprintf(msg, args...), cause)
	case sarama.ErrClosedClient, sarama.ErrAlreadyConnected,
		sarama.ErrNotConnected, sarama.ErrShuttingDown, sarama.ErrControllerNotAvailable:
		return NewKafkaError(ErrorCodeIllegalState, fmt.Sprintf(msg, args...), cause)
	case sarama.ErrInvalidPartition, sarama.ErrIncompleteResponse,
		sarama.ErrInsufficientData, sarama.ErrMessageTooLarge, sarama.ErrNoTopicsToUpdateMetadata:
		return ErrorSubTypeProvisioning.WithCause(cause, msg, args...)
	case sarama.ErrConsumerOffsetNotAdvanced:
		// note, this should not happen during binding, we use generic internal
		fallthrough
	default:
		return NewKafkaError(ErrorCodeBindingInternal, fmt.Sprintf(msg, args...), cause)
	}
}
