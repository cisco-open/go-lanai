package kafka

import (
	"encoding"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
)

type jsonEncoder struct{}

func (enc jsonEncoder) MIMEType() string {
	return MIMETypeJson
}

func (enc jsonEncoder) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

type binaryEncoder struct{}

func (enc binaryEncoder) MIMEType() string {
	return MIMETypeBinary
}

func (enc binaryEncoder) Encode(v interface{}) ([]byte, error) {
	switch key := v.(type) {
	case string:
		return []byte(key), nil
	case []byte:
		return key, nil
	case encoding.BinaryMarshaler:
		return key.MarshalBinary()
	default:
		return nil, fmt.Errorf("unsupported value for binary encoding: %T", v)
	}
}

type saramaEncoderWrapper struct {
	v     interface{}
	enc   Encoder
	cache []byte
}

func newSaramaEncoder(v interface{}, enc Encoder) sarama.Encoder {
	return &saramaEncoderWrapper{
		v:     v,
		enc:   enc,
	}
}

func (w saramaEncoderWrapper) Encode() (ret []byte, err error) {
	if w.cache != nil {
		return w.cache, nil
	}
	defer func() {
		w.cache = ret
	}()
	ret, err = w.enc.Encode(w.v)
	return
}

func (w saramaEncoderWrapper) Length() int {
	data, e := w.Encode()
	if e != nil {
		return 0
	}
	return len(data)
}

// mimeTypeProducerInterceptor implement ProducerInterceptor.
// This interceptor applies value encoder and set Content-Type to message headers
type mimeTypeProducerInterceptor struct{}

func (i mimeTypeProducerInterceptor) Intercept(msgCtx *MessageContext) (*MessageContext, error) {
	if msgCtx.Headers == nil {
		msgCtx.Headers = Headers{}
	}

	if msgCtx.ValueEncoder == nil {
		msgCtx.ValueEncoder = jsonEncoder{}
	}

	msgCtx.Headers[HeaderContentType] = msgCtx.ValueEncoder.MIMEType()
	msgCtx.Payload = newSaramaEncoder(msgCtx.Payload, msgCtx.ValueEncoder)
	return msgCtx, nil
}



