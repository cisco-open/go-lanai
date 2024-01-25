package log

import (
	"io"
	"strings"
)

// writerAdapter implements io.Writer and wrap around our Logger interface
type writerAdapter struct {
	logger Logger
}

func NewWriterAdapter(logger Logger, lvl LoggingLevel) io.Writer {
	return &writerAdapter{
		logger: logger.WithLevel(lvl),
	}
}

func (w writerAdapter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	w.logger.Print(strings.TrimSpace(string(p)))
	return len(p), nil
}
