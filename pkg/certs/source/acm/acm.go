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

package acmcerts

import (
    "context"
    "crypto/ecdsa"
    "crypto/rsa"
    "crypto/tls"
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/acm"
    "github.com/cisco-open/go-lanai/pkg/certs"
    certsource "github.com/cisco-open/go-lanai/pkg/certs/source"
    "github.com/cisco-open/go-lanai/pkg/utils/loop"
    "go.step.sm/crypto/pemutil"
    "regexp"
    "strings"
    "sync"
    "time"
)

type AcmProvider struct {
	props             SourceProperties
	acmClient         *acm.Client
	cache             *certsource.FileCache
	cachedCertificate *tls.Certificate
	lcCtx             context.Context
	mutex             sync.RWMutex
	once              sync.Once
	monitor           *loop.Loop
	monitorCancel     context.CancelFunc
}

func NewAcmProvider(ctx context.Context, acm *acm.Client, p SourceProperties) certs.Source {
	cache, e := certsource.NewFileCache(func(opt *certsource.FileCacheOption) {
		opt.Root = p.CachePath
		opt.Type = sourceType
		opt.Prefix = resolveCacheKey(&p)
	})
	if e != nil {
		logger.WithContext(ctx).Warnf("file cache for %s certificate source is not enabled: %v", sourceType, e)
	}
	return &AcmProvider{
		props:     p,
		acmClient: acm,
		cache:     cache,
		lcCtx:     ctx,
		monitor:   loop.NewLoop(),
	}
}

func (a *AcmProvider) Close() error {
	return nil
}

func (a *AcmProvider) TLSConfig(ctx context.Context, _ ...certs.TLSOptions) (*tls.Config, error) {
	if e := a.LazyInit(ctx); e != nil {
		return nil, e
	}
	rootCAs, e := a.RootCAs(ctx)
	if e != nil {
		return nil, e
	}
	minVer, e := certsource.ParseTLSVersion(a.props.MinTLSVersion)
	if e != nil {
		return nil, e
	}
	//nolint:gosec // false positive -  G402: TLS MinVersion too low
	return &tls.Config{
		GetClientCertificate: a.toGetClientCertificateFunc(),
		RootCAs:              rootCAs,
		MinVersion:           minVer,
	}, nil
}

func (a *AcmProvider) Files(ctx context.Context) (*certs.CertificateFiles, error) {
	if e := a.LazyInit(ctx); e != nil {
		return nil, e
	}
	if a.cache == nil {
		return nil, fmt.Errorf("unable to access certificates as local files: file cache is not enabled for source [%s]", sourceType)
	}

	return &certs.CertificateFiles{
		RootCAPaths:     []string{a.cache.ResolvePath(certsource.CachedFileKeyCA)},
		CertificatePath: a.cache.ResolvePath(certsource.CachedFileKeyCertificate),
		PrivateKeyPath:  a.cache.ResolvePath(certsource.CachedFileKeyPrivateKey),
	}, nil
}

func (a *AcmProvider) RootCAs(ctx context.Context) (*x509.CertPool, error) {
	input := &acm.ExportCertificateInput{
		CertificateArn: &a.props.ARN,
		Passphrase:     []byte(a.props.Passphrase),
	}
	output, err := a.acmClient.ExportCertificate(ctx, input)
	if err != nil {
		logger.Errorf("Could not fetch ACM certificate %s: %s", a.props.ARN, err.Error())
		return nil, err
	}
	//Clean the returned CA (deal with bug in localStack)
	cleantext := strings.Replace(*output.CertificateChain, " -----END CERTIFICATE-----", "-----END CERTIFICATE-----", -1)
	pemBytes := []byte(cleantext)
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(pemBytes)
	if a.cache != nil {
		if err := a.cache.CachePEM(pemBytes, certsource.CachedFileKeyCA); err != nil {
			logger.WithContext(ctx).Warnf(`unable to cache CA: %v`, err)
			return certPool, err
		}
	}
	return certPool, nil
}

func (a *AcmProvider) LazyInit(ctx context.Context) error {
	var err error
	//nolint:contextcheck // false positive, sync.Once.Do doesn't take func(context.Context)
	a.once.Do(func() {
		// At least get RootCA once
		// TODO should we renew RootCA periodically?
		_, err = a.RootCAs(ctx)
		if err != nil {
			logger.Errorf("Failed to get CAs from ACM: %s", err.Error())
			return
		}
		// At least get Certificate once
		var cert *tls.Certificate
		cert, err = a.generateClientCertificate(ctx)
		if err != nil {
			logger.Errorf("Failed to get certificate from ACM: %s", err.Error())
			return
		}
		renewIntervalFunc := certsource.RenewRepeatIntervalFunc(time.Duration(a.props.MinRenewInterval))
		delay := renewIntervalFunc(cert, err)

		loopCtx, cancelFunc := a.monitor.Run(a.lcCtx)
		a.monitorCancel = cancelFunc

		time.AfterFunc(delay, func() {
			a.monitor.Repeat(a.tryRenew(loopCtx), func(opt *loop.TaskOption) {
				opt.RepeatIntervalFunc = renewIntervalFunc
			})
		})
	})
	return err
}

func (a *AcmProvider) toGetClientCertificateFunc() func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
	return func(certificateReq *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		a.mutex.RLock()
		defer a.mutex.RUnlock()
		if a.cachedCertificate == nil {
			return new(tls.Certificate), nil
		}
		e := certificateReq.SupportsCertificate(a.cachedCertificate)
		if e != nil {
			// No acceptable certificate found. Don't send a certificate. Don't need to treat as error.
			// see tls package's func (c *Conn) getClientCertificate(cri *CertificateRequestInfo) (*Certificate, error)
			return new(tls.Certificate), nil //nolint:nilerr // as intended
		} else {
			return a.cachedCertificate, nil
		}
	}
}

func (a *AcmProvider) generateClientCertificate(ctx context.Context) (*tls.Certificate, error) {
	input := &acm.ExportCertificateInput{
		CertificateArn: &a.props.ARN,
		Passphrase:     []byte(a.props.Passphrase),
	}
	output, err := a.acmClient.ExportCertificate(ctx, input)
	if err != nil {
		logger.Errorf("Could not fetch ACM certificate %s: %s", a.props.ARN, err.Error())
		return nil, err
	}
	crtPEM := []byte(*output.Certificate)

	keyBlock, _ := pem.Decode([]byte(*output.PrivateKey))
	unEncryptedKey, err := pemutil.DecryptPKCS8PrivateKey(keyBlock.Bytes, []byte(a.props.Passphrase))
	if err != nil {
		logger.Errorf("Could not decrypt pkcs8 private key: %s", err.Error())
		return nil, err
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(unEncryptedKey)
	if err != nil {
		logger.Errorf("Could not parse pkcs8 private key: %s", err.Error())
		return nil, err
	}
	var privDER []byte
	switch privateKey.(type) {
	case *rsa.PrivateKey:
		privDER = x509.MarshalPKCS1PrivateKey(privateKey.(*rsa.PrivateKey))
	case *ecdsa.PrivateKey:
		privDER, err = x509.MarshalECPrivateKey(privateKey.(*ecdsa.PrivateKey))
		if err != nil {
			logger.Errorf("Could not marshal ecdsa private key: %s", err.Error())
			return nil, err
		}

	default:
		panic("unknown key")
	}
	keyBlock.Bytes = privDER
	keyBlock.Headers = nil
	keyBytes := pem.EncodeToMemory(keyBlock)

	cert, err := tls.X509KeyPair(crtPEM, keyBytes)
	if err != nil {
		logger.Errorf("Could not create cert from PEM: %s", err.Error())
		return nil, err
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.cachedCertificate = &cert
	if a.cache != nil {
		if err := a.cache.CacheCertificate(&cert); err != nil {
			logger.WithContext(ctx).Warnf(`unable to cache certificate: %v`, err)
			return &cert, err
		}
	}
	return &cert, nil
}

func (a *AcmProvider) renewClientCertificate(ctx context.Context) error {
	input := &acm.RenewCertificateInput{
		CertificateArn: &a.props.ARN,
	}
	_, err := a.acmClient.RenewCertificate(ctx, input)
	if err != nil {
		logger.Errorf("Could not renew ACM certificate %s: %s", a.props.ARN, err.Error())
		return err
	}
	return nil
}

func (a *AcmProvider) tryRenew(loopCtx context.Context) loop.TaskFunc {
	return func(_ context.Context, l *loop.Loop) (ret interface{}, err error) {
		//ignore error since we will just schedule another renew
		err = a.renewClientCertificate(loopCtx)
		if err != nil {
			logger.Warn("certificate renew failed: %v", err)
		}
		ret, err = a.generateClientCertificate(loopCtx)
		if err != nil {
			logger.Warn("certificate renew failed: %v", err)
		} else {
			logger.Infof("certificate has been renewed")
		}
		return
	}
}

var (
	arnRegex         = regexp.MustCompile(`arn:(?P<part>[^:]+):(?P<srv>[^:]+):(?P<region>[^:]+):(?P<acct>[^:]+):((?P<res_type>[^:]+)[\/:])?(?P<res_id>[^:]+)$`)
	cacheKeyReplacer = strings.NewReplacer(
		" ", "-",
		".", "-",
		"_", "-",
		"@", "-at-",
	)
	cacheKeyCount = 0
)

func resolveCacheKey(p *SourceProperties) (key string) {
	var resId, resType string
	matches := arnRegex.FindStringSubmatch(p.ARN)
	for i, name := range arnRegex.SubexpNames() {
		if i >= len(matches) {
			break
		}
		switch name {
		case "res_id":
			resId = matches[i]
		case "res_type":
			resType = matches[i]
		}
	}
	cacheKeyCount++
	key = fmt.Sprintf(`%s-%s-%d`, resType, resId, cacheKeyCount)
	key = cacheKeyReplacer.Replace(key)
	return key
}
