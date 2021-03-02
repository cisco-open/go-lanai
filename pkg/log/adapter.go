package log

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"strings"
)

// writerAdapter implements io.Writer and wrap around our Logger interface
type writerAdapter struct {
	logger log.Logger
}

func NewWriterAdapter(logger Logger, lvl LoggingLevel) *writerAdapter {
	kitLogger := log.Logger(logger)
	switch lvl {
	case LevelDebug:
		kitLogger = level.Debug(logger)
	case LevelInfo:
		kitLogger = level.Info(logger)
	case LevelWarn:
		kitLogger = level.Warn(logger)
	case LevelError:
		kitLogger = level.Error(logger)
	}
	return &writerAdapter{
		logger: kitLogger,
	}
}

func (w writerAdapter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	msg := strings.TrimSpace(string(p))
	return len(p), w.logger.Log(LogKeyMessage, msg)
}
