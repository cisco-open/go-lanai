package kafka

import (
	"encoding"
	"encoding/json"
	"github.com/Shopify/sarama"
)

type jsonEncoder struct{}

func (enc jsonEncoder) MIMEType() string {
	return MIMETypeJson
}

func (enc jsonEncoder) Encode(v interface{}) (bytes []byte, err error) {
	if bytes, err = json.Marshal(v); err != nil {
		return bytes, ErrorSubTypeEncoding.WithCause(err, err.Error())
	}
	return
}

type binaryEncoder struct{}

func (enc binaryEncoder) MIMEType() string {
	return MIMETypeBinary
}

func (enc binaryEncoder) Encode(v interface{}) (bytes []byte, err error) {
	switch val := v.(type) {
	case string:
		return []byte(val), nil
	case []byte:
		return val, nil
	case encoding.BinaryMarshaler:
		if bytes, err = val.MarshalBinary(); err != nil {
			return bytes, ErrorSubTypeEncoding.WithCause(err, err.Error())
		}
		return
	default:
		return nil, ErrorSubTypeEncoding.WithMessage("unsupported value for binary encoding: %T", v)
	}
}

type saramaEncoderWrapper struct {
	v     interface{}
	enc   Encoder
	cache []byte
}

func newSaramaEncoder(v interface{}, enc Encoder) sarama.Encoder {
	if v == nil {
		return nil
	}
	if enc == nil {
		enc = binaryEncoder{}
	}
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

// mimeTypeProducerInterceptor implement ProducerMessageInterceptor.
// This interceptor applies value encoder and set Content-Type to message headers
type mimeTypeProducerInterceptor struct{}

func (i mimeTypeProducerInterceptor) Intercept(msgCtx *MessageContext) (*MessageContext, error) {
	if msgCtx.Message.Headers == nil {
		msgCtx.Message.Headers = Headers{}
	}

	if msgCtx.ValueEncoder == nil {
		msgCtx.ValueEncoder = jsonEncoder{}
	}

	msgCtx.Message.Headers[HeaderContentType] = msgCtx.ValueEncoder.MIMEType()
	msgCtx.Message.Payload = newSaramaEncoder(msgCtx.Message.Payload, msgCtx.ValueEncoder)
	return msgCtx, nil
}



