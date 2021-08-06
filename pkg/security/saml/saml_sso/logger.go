package saml_auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
	"os"
)

type loggerAdapter struct {
	delegate log.Logger
}

func newLoggerAdaptor(l log.Logger) *loggerAdapter{
	return &loggerAdapter{
		delegate: l,
	}
}

func (s *loggerAdapter) Printf(format string, v ...interface{}) {
	s.delegate.Info(format, v...)
}

func (s *loggerAdapter) Print(v ...interface{}) {
	s.delegate.Info("", v...)
}

func (s *loggerAdapter) Println(v ...interface{}) {
	s.delegate.Info("", v...)
}

func (s *loggerAdapter) Fatal(v ...interface{}) {
	s.delegate.Error("", v...)
	os.Exit(1)
}

func (s *loggerAdapter) Fatalf(format string, v ...interface{}) {
	s.delegate.Errorf("", v...)
	os.Exit(1)
}

func (s *loggerAdapter) Fatalln(v ...interface{}) {
	s.delegate.Error("", v...)
	os.Exit(1)
}

func (s *loggerAdapter) Panic(v ...interface{}) {
	s.delegate.Error("", v...)
	panic(fmt.Sprint(v...))
}

func (s *loggerAdapter) Panicf(format string, v ...interface{}) {
	s.delegate.Errorf(format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (s *loggerAdapter) Panicln(v ...interface{}) {
	s.delegate.Error("", v...)
	panic(fmt.Sprint(v...))
}
