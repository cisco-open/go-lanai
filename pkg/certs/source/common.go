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

package certsource

import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "dario.cat/mergo"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/certs"
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/cisco-open/go-lanai/pkg/utils/loop"
    "time"
)

var tlsVersions = map[string]uint16{
	"":      tls.VersionTLS10, // default in golang
	"tls10": tls.VersionTLS10,
	"tls11": tls.VersionTLS11,
	"tls12": tls.VersionTLS12,
	"tls13": tls.VersionTLS13,
}

var logger = log.New("Certs.Source")

func NewFactory[PropertiesType any](typ certs.SourceType, rawDefaultConfig json.RawMessage, constructor func(props PropertiesType) certs.Source) (*GenericFactory[PropertiesType], error) {
	var zero PropertiesType
	defaults, e := ParseConfigWithDefaults(zero, rawDefaultConfig)
	if e != nil {
		return nil, fmt.Errorf(`unable to parse default certificate source configuration with type [%s]: %v`, typ, e)
	}
	if constructor == nil {
		return nil, fmt.Errorf(`constructor of certificate source with type [%s] is missing`, typ)
	}
	return &GenericFactory[PropertiesType]{
		SourceType:  typ,
		Defaults:    defaults,
		Constructor: constructor,
	}, nil
}

type GenericFactory[PropertiesType any] struct {
	SourceType  certs.SourceType
	Defaults    PropertiesType
	Constructor func(props PropertiesType) certs.Source
}

func (f *GenericFactory[T]) Type() certs.SourceType {
	return f.SourceType
}

func (f *GenericFactory[T]) LoadAndInit(_ context.Context, opts ...certs.SourceOptions) (certs.Source, error) {
	src := certs.SourceConfig{}
	for _, fn := range opts {
		fn(&src)
	}
	props, e := ParseConfigWithDefaults(f.Defaults, src.RawConfig)
	if e != nil {
		return nil, fmt.Errorf(`unable to initialize certificate source [%s]: %v`, f.Type(), e)
	}
	return f.Constructor(props), nil
}

func ParseConfigWithDefaults[T any](defaults T, rawConfig json.RawMessage) (T, error) {
	if rawConfig == nil || len(rawConfig) == 0 {
		return defaults, nil
	}

	var parsed T
	if e := json.Unmarshal(rawConfig, &parsed); e != nil {
		return defaults, e
	}

	if e := mergo.Merge(&defaults, &parsed, mergo.WithOverride); e != nil {
		return defaults, e
	}
	return defaults, nil
}

func ParseTLSVersion(verStr string) (uint16, error) {
	if v, ok := tlsVersions[verStr]; ok {
		return v, nil
	} else {
		return tls.VersionTLS10, errors.New(fmt.Sprintf("unsupported tls version %s", verStr))
	}
}

// RenewRepeatIntervalFunc create a loop.RepeatIntervalFunc for renewing certificate.
// The interval is set to the half way between now and cached certificate expiration.
// If "fallbackInterval" is provided, it is used for any error cases
func RenewRepeatIntervalFunc(fallbackInterval time.Duration) loop.RepeatIntervalFunc {
	return func(result interface{}, err error) (ret time.Duration) {
		defer func() {
			logger.Debugf("certificate will renew in %v", ret)
		}()

		minDuration := 1 * time.Minute
		if fallbackInterval  != 0 {
			minDuration = fallbackInterval
		}

		if err != nil {
			return minDuration
		}

		cert := result.(*tls.Certificate)
		if len(cert.Certificate) < 1 {
			return minDuration
		}

		parsedCert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return minDuration
		}

		validTo := parsedCert.NotAfter
		now := time.Now()

		if validTo.Before(now) {
			return minDuration
		}

		durationRemain := validTo.Sub(now)
		next := durationRemain / 2

		if minDuration > next {
			next = minDuration
		}
		return next
	}
}
