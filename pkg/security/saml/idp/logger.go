// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package samlidp

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
	s.delegate.Infof(format, v...)
}

func (s *loggerAdapter) Print(v ...interface{}) {
	s.delegate.Info(fmt.Sprint(v...))
}

func (s *loggerAdapter) Println(v ...interface{}) {
	s.Print(v...)
}

func (s *loggerAdapter) Fatal(v ...interface{}) {
	s.delegate.Error(fmt.Sprint(v...))
	os.Exit(1)
}

func (s *loggerAdapter) Fatalf(format string, v ...interface{}) {
	s.delegate.Errorf(format, v...)
	os.Exit(1)
}

func (s *loggerAdapter) Fatalln(v ...interface{}) {
	s.Fatal(v...)
}

func (s *loggerAdapter) Panic(v ...interface{}) {
	s.delegate.Error(fmt.Sprint(v...))
	panic(fmt.Sprint(v...))
}

func (s *loggerAdapter) Panicf(format string, v ...interface{}) {
	s.delegate.Errorf(format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (s *loggerAdapter) Panicln(v ...interface{}) {
	s.Panic(v...)
}
