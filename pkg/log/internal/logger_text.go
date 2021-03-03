package internal

import (
	"errors"
	"io"
)

var errMissingValue = errors.New("(MISSING)")

type Fields map[string]interface{}

// KitTextLoggerAdapter implmenets go-kit's log.Logger with custom Formatter
type KitTextLoggerAdapter struct {
	Formatter  TextFormatter
	Writer     io.Writer
	IsTerminal bool
}

func NewKitTextLoggerAdapter(writer io.Writer, formatter TextFormatter) *KitTextLoggerAdapter {
	return &KitTextLoggerAdapter{
		Formatter: formatter,
		Writer: writer,
		IsTerminal: IsTerminal(writer),
	}
}

func (l *KitTextLoggerAdapter) Log(keyvals ...interface{}) error {
	values := Fields{}
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			values[Sprint(keyvals[i])] = keyvals[i+1]
		} else {
			values[Sprint(keyvals[i])] = errMissingValue
		}
	}
	return l.Formatter.Format(values, l.Writer)
}


