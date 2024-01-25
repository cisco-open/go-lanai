package kafka

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
	"github.com/IBM/sarama"
)

type MessageLogger interface {
	WithLevel(level log.LoggingLevel) MessageLogger
	LogSentMessage(ctx context.Context, msg interface{})
	LogReceivedMessage(ctx context.Context, msg interface{})
}

type LoggerOptions func(opt *loggerOption)

type loggerOption struct {
	Name  string
	Level log.LoggingLevel
}

type saramaMessageLogger struct {
	logger log.ContextualLogger
	level  log.LoggingLevel
}

func newSaramaMessageLogger(opts ...LoggerOptions) *saramaMessageLogger {
	opt := loggerOption{
		Name:  "Kafka.Msg",
		Level: log.LevelDebug,
	}
	for _, fn := range opts {
		fn(&opt)
	}
	return &saramaMessageLogger{
		logger: log.New(opt.Name),
		level:  opt.Level,
	}
}

func (l saramaMessageLogger) WithLevel(level log.LoggingLevel) MessageLogger {
	return &saramaMessageLogger{
		logger: l.logger,
		level:  level,
	}
}

func (l saramaMessageLogger) LogSentMessage(ctx context.Context, msg interface{}) {
	switch m := msg.(type) {
	case *sarama.ProducerMessage:
		logMsg := fmt.Sprintf("[SENT] [%s] Partition[%d] Offset[%d]: Length=%dB",
			m.Topic, m.Partition, m.Offset, m.Value.Length())
		if m.Key != nil && m.Key.Length() != 0 {
			logMsg = logMsg + fmt.Sprintf(" KeyLength=%dB", m.Key.Length())
		}
		logger.WithContext(ctx).WithLevel(l.level).Printf(logMsg)
	}
}

func (l saramaMessageLogger) LogReceivedMessage(ctx context.Context, msg interface{}) {
	switch m := msg.(type) {
	case *sarama.ConsumerMessage:
		logMsg := fmt.Sprintf("[RECV] [%s] Partition[%d] Offset[%d]: Length=%dB",
			m.Topic, m.Partition, m.Offset, len(m.Value))
		if len(m.Key) != 0 {
			logMsg = logMsg + fmt.Sprintf(" Key=%x", m.Key)
		}
		logger.WithContext(ctx).WithLevel(l.level).Printf(logMsg)
	}
}
