package acmcerts

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	certsource "cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs/source"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/loop"
	"encoding/pem"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	awsacm "github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
	"go.step.sm/crypto/pemutil"
	"regexp"
	"strings"
	"sync"
	"time"
)

type AcmProvider struct {
	props             SourceProperties
	acmClient         acmiface.ACMAPI
	cache             *certsource.FileCache
	cachedCertificate *tls.Certificate
	mutex             sync.RWMutex
	once              sync.Once
	monitor           *loop.Loop
	monitorCancel     context.CancelFunc
}

func NewAcmProvider(acm acmiface.ACMAPI, p SourceProperties) certs.Source {
	cache, e := certsource.NewFileCache(func(opt *certsource.FileCacheOption) {
		opt.Root = p.CachePath
		opt.Type = sourceType
		opt.Prefix = resolveCacheKey(&p)
	})
	if e != nil {
		logger.Warnf("file cache for %s certificate source is not enabled: %v", sourceType, e)
	}
	return &AcmProvider{
		props:     p,
		acmClient: acm,
		cache:     cache,
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

// GetMinTlsVersion
// Deprecated
func (a *AcmProvider) GetMinTlsVersion() (uint16, error) {
	return certsource.ParseTLSVersion(a.props.MinTLSVersion)
}

func (a *AcmProvider) RootCAs(ctx context.Context) (*x509.CertPool, error) {
	input := &awsacm.ExportCertificateInput{
		CertificateArn: aws.String(a.props.ARN),
		Passphrase:     []byte(a.props.Passphrase),
	}
	output, err := a.acmClient.ExportCertificateWithContext(ctx, input)
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

// GetClientCertificate
// Deprecated
func (a *AcmProvider) GetClientCertificate(ctx context.Context) (func(*tls.CertificateRequestInfo) (*tls.Certificate, error), error) {
	if e := a.LazyInit(ctx); e != nil {
		return nil, e
	}
	return a.toGetClientCertificateFunc(), nil

}

func (a *AcmProvider) LazyInit(ctx context.Context) error {
	var err error
	a.once.Do(func() {
		var cert *tls.Certificate
		cert, err = a.generateClientCertificate(ctx)
		if err != nil {
			logger.Errorf("Failed to get certificate from ACM: %s", err.Error())
			return
		}
		renewIntervalFunc := certsource.RenewRepeatIntervalFunc(time.Duration(a.props.MinRenewInterval))
		delay := renewIntervalFunc(cert, err)

		loopCtx, cancelFunc := a.monitor.Run(context.Background())
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
			return new(tls.Certificate), nil //nolint:nilerr
		} else {
			return a.cachedCertificate, nil
		}
	}
}

func (a *AcmProvider) generateClientCertificate(ctx context.Context) (*tls.Certificate, error) {
	input := &awsacm.ExportCertificateInput{
		CertificateArn: aws.String(a.props.ARN),
		Passphrase:     []byte(a.props.Passphrase),
	}
	output, err := a.acmClient.ExportCertificateWithContext(ctx, input)
	if err != nil {
		logger.Errorf("Could not fetch ACM certificate %s: %s", a.props.ARN, err.Error())
		return nil, err
	}
	crtPEM := []byte(*output.Certificate)

	keyBlock, _ := pem.Decode([]byte(*output.PrivateKey))
	//nolint:staticcheck
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

func (a *AcmProvider) renewClientCertificate(_ context.Context) error {
	input := &awsacm.RenewCertificateInput{
		CertificateArn: aws.String(a.props.ARN),
	}
	_, err := a.acmClient.RenewCertificate(input)
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
