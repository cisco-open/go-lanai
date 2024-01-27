package internal

import (
	"go.uber.org/zap/zapcore"
	"io"
)

type TerminalAware interface {
	IsTerminal() bool
}

// ZapWriterWrapper implements zapcore.WriteSyncer and TerminalAware
type ZapWriterWrapper struct {
	io.Writer
}

func (ZapWriterWrapper) Sync() error {
	return nil
}

func (s ZapWriterWrapper) IsTerminal() bool {
	return IsTerminal(s.Writer)
}

// NewZapWriterWrapper similar to zapcore.AddSync with exported type
func NewZapWriterWrapper(w io.Writer) zapcore.WriteSyncer {
	return ZapWriterWrapper{
		Writer: w,
	}
}

// ZapTerminalCore implements TerminalAware and always returns true
type ZapTerminalCore struct {
	zapcore.Core
}

func (s ZapTerminalCore) IsTerminal() bool {
	return true
}