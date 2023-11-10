package tlsconfig

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/aws/acm"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/loop"
	"encoding/pem"
	"github.com/aws/aws-sdk-go/aws"
	awsacm "github.com/aws/aws-sdk-go/service/acm"
	"go.step.sm/crypto/pemutil"
	"strings"
	"sync"
	"time"
)

type AcmProvider struct {
	ProviderCommon
	acmClient         *acm.AcmClient
	cachedCertificate *tls.Certificate
	mutex             sync.RWMutex
	once              sync.Once
	monitor           *loop.Loop
	monitorCancel     context.CancelFunc
}

func NewAcmProvider(acm *acm.AcmClient, p Properties) *AcmProvider {
	return &AcmProvider{
		ProviderCommon: ProviderCommon{
			p,
		},
		acmClient: acm,
		monitor:   loop.NewLoop(),
	}
}

func (a *AcmProvider) Close() error {
	return nil
}

func (a *AcmProvider) RootCAs(ctx context.Context) (*x509.CertPool, error) {
	input := &awsacm.ExportCertificateInput{
		CertificateArn: aws.String(a.p.Arn),
		Passphrase:     []byte(a.p.Passphrase),
	}
	output, err := a.acmClient.Client.ExportCertificateWithContext(ctx, input)
	if err != nil {
		logger.Errorf("Could not fetch ACM certificate %s: %s", a.p.Arn, err.Error())
		return nil, err
	}
	//Clean the returned CA (deal with bug in localStack)
	cleantext := strings.Replace(*output.CertificateChain, " -----END CERTIFICATE-----", "-----END CERTIFICATE-----", -1)
	pemBytes := []byte(cleantext)
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(pemBytes)
	if a.p.FileCache.Enabled {
		err := a.CacheCaToFile(pemBytes)
		if err != nil {
			return certPool, err
		}
	}
	return certPool, nil
}

func (a *AcmProvider) GetClientCertificate(ctx context.Context) (func(*tls.CertificateRequestInfo) (*tls.Certificate, error), error) {
	a.once.Do(func() {
		cert, err := a.generateClientCertificate(ctx)
		if err != nil {
			logger.Errorf("Failed to get certificate from ACM: %s", err.Error())
			return
		}
		delay := a.tryRenewRepeatIntervalFunc()(cert, err)

		loopCtx, cancelFunc := a.monitor.Run(context.Background())
		a.monitorCancel = cancelFunc

		time.AfterFunc(delay, func() {
			a.monitor.Repeat(a.tryRenew(loopCtx), func(opt *loop.TaskOption) {
				opt.RepeatIntervalFunc = a.tryRenewRepeatIntervalFunc()
			})
		})
	})
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
	}, nil

}

func (a *AcmProvider) generateClientCertificate(ctx context.Context) (*tls.Certificate, error) {
	input := &awsacm.ExportCertificateInput{
		CertificateArn: aws.String(a.p.Arn),
		Passphrase:     []byte(a.p.Passphrase),
	}
	output, err := a.acmClient.Client.ExportCertificateWithContext(ctx, input)
	if err != nil {
		logger.Errorf("Could not fetch ACM certificate %s: %s", a.p.Arn, err.Error())
		return nil, err
	}
	crtPEM := []byte(*output.Certificate)

	keyBlock, _ := pem.Decode([]byte(*output.PrivateKey))
	//nolint:staticcheck
	unEncryptedKey, err := pemutil.DecryptPKCS8PrivateKey(keyBlock.Bytes, []byte(a.p.Passphrase))
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

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.cachedCertificate = &cert
	if a.p.FileCache.Enabled {
		err := a.CacheCertToFile(&cert)
		if err != nil {
			return &cert, err
		}
	}
	return &cert, nil
}

func (a *AcmProvider) renewClientCertificate(ctx context.Context) error {
	input := &awsacm.RenewCertificateInput{
		CertificateArn: aws.String(a.p.Arn),
	}
	_, err := a.acmClient.Client.RenewCertificate(input)
	if err != nil {
		logger.Errorf("Could not renew ACM certificate %s: %s", a.p.Arn, err.Error())
		return err
	}
	return nil
}

// CacheCertToFile will write out a cert and key to files based on configured path and prefix
func (a *AcmProvider) CacheCertToFile(cert *tls.Certificate) error {
	certfilepath := a.p.FileCache.Path + a.ProviderCommon.p.FileCache.Prefix + CertSuffix
	keyfilepath := a.p.FileCache.Path + a.ProviderCommon.p.FileCache.Prefix + KeySuffix
	return CacheCertToFile(cert, certfilepath, keyfilepath)
}

// CacheCaToFile writes the provided ca cert pool to a file based on the provided config
func (a AcmProvider) CacheCaToFile(pemData []byte) error {
	cafilepath := a.p.FileCache.Path + a.ProviderCommon.p.FileCache.Prefix + CaSuffix
	return CacheCaToFile(pemData, cafilepath)
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
