package vaultcerts

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	certsource "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig/source"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/loop"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"io"
	"path"
	"sync"
	"time"
)

type VaultProvider struct {
	p SourceProperties

	vc *vault.Client

	once              sync.Once
	mutex             sync.RWMutex
	cachedCertificate *tls.Certificate

	lcCtx         context.Context
	monitor       *loop.Loop
	monitorCancel context.CancelFunc
}

func NewVaultProvider(lcCtx context.Context, vc *vault.Client, p SourceProperties) tlsconfig.Source {
	if lcCtx == nil {
		lcCtx = context.Background()
	}
	return &VaultProvider{
		p:       p,
		vc:      vc,
		lcCtx:   lcCtx,
		monitor: loop.NewLoop(),
	}
}

func (v *VaultProvider) TLSConfig(ctx context.Context, _ ...tlsconfig.TLSOptions) (*tls.Config, error) {
	if e := v.LazyInit(ctx); e != nil {
		return nil, e
	}
	rootCAs, e := v.RootCAs(ctx)
	if e != nil {
		return nil, e
	}
	minVer, e := certsource.ParseTLSVersion(v.p.MinTLSVersion)
	if e != nil {
		return nil, e
	}
	return &tls.Config{
		GetClientCertificate: v.toGetClientCertificateFunc(),
		RootCAs:              rootCAs,
		MinVersion:           minVer,
	}, nil
}

func (v *VaultProvider) Files(ctx context.Context) (*tlsconfig.CertificateFiles, error) {
	if e := v.LazyInit(ctx); e != nil {
		return nil, e
	}

	// TODO generalize this impl
	cafilepath := v.p.FileCache.Path + v.p.FileCache.Prefix + certsource.CaSuffix
	certfilepath := v.p.FileCache.Path + v.p.FileCache.Prefix + certsource.CertSuffix
	keyfilepath := v.p.FileCache.Path + v.p.FileCache.Prefix + certsource.KeySuffix
	return &tlsconfig.CertificateFiles{
		RootCAPaths:          []string{cafilepath},
		CertificatePath:      certfilepath,
		PrivateKeyPath:       keyfilepath,
	}, nil
}

func (v *VaultProvider) GetClientCertificate(ctx context.Context) (func(*tls.CertificateRequestInfo) (*tls.Certificate, error), error) {
	if e := v.LazyInit(ctx); e != nil {
		return nil, e
	}
	return v.toGetClientCertificateFunc(), nil
}

func (v *VaultProvider) RootCAs(ctx context.Context) (*x509.CertPool, error) {
	resp, err := v.vc.Logical(ctx).ReadRawWithContext(ctx, path.Join(v.p.Path, "ca", "pem"))
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	pemBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(pemBytes)
	if v.p.FileCache.Enabled {
		logger.WithContext(ctx).Infof("gonna cache the ca using config: %v", v.p.FileCache)
		err := v.CacheCaToFile(pemBytes)
		if err != nil {
			return certPool, err
		}
	}
	return certPool, nil
}

// GetMinTlsVersion
// Deprecated
func (v *VaultProvider) GetMinTlsVersion() (uint16, error) {
	return certsource.ParseTLSVersion(v.p.MinTLSVersion)
}

func (v *VaultProvider) LazyInit(ctx context.Context) error {
	var err error
	v.once.Do(func() {
		var cert *tls.Certificate
		cert, err = v.generateClientCertificate(ctx)
		if err != nil {
			logger.Errorf("Failed to get certificate from Vault: %s", err.Error())
			return
		}
		renewIntervalFunc := certsource.RenewRepeatIntervalFunc(time.Duration(v.p.MinRenewInterval))
		delay := renewIntervalFunc(cert, err)

		loopCtx, cancelFunc := v.monitor.Run(v.lcCtx)
		v.monitorCancel = cancelFunc

		time.AfterFunc(delay, func() {
			v.monitor.Repeat(v.tryRenew(loopCtx), func(opt *loop.TaskOption) {
				opt.RepeatIntervalFunc = renewIntervalFunc
			})
		})
	})
	return err
}

func (v *VaultProvider) Close() error {
	if v.monitorCancel != nil {
		v.monitorCancel()
	}
	return nil
}

func (v *VaultProvider) toGetClientCertificateFunc() func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
	return func(certificateReq *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		v.mutex.RLock()
		defer v.mutex.RUnlock()
		if v.cachedCertificate == nil {
			return new(tls.Certificate), nil
		}
		e := certificateReq.SupportsCertificate(v.cachedCertificate)
		if e != nil {
			// No acceptable certificate found. Don't send a certificate. Don't need to treat as error.
			// see tls package's func (c *Conn) getClientCertificate(cri *CertificateRequestInfo) (*Certificate, error)
			return new(tls.Certificate), nil //nolint:nilerr
		} else {
			return v.cachedCertificate, nil
		}
	}
}

func (v *VaultProvider) tryRenew(loopCtx context.Context) loop.TaskFunc {
	return func(_ context.Context, l *loop.Loop) (ret interface{}, err error) {
		//ignore error since we will just schedule another renew
		ret, err = v.generateClientCertificate(loopCtx)
		if err != nil {
			logger.Warn("certificate renew failed: %v", err)
		} else {
			logger.Infof("certificate has been renewed")
		}
		return
	}
}

func (v *VaultProvider) generateClientCertificate(ctx context.Context) (*tls.Certificate, error) {
	fullPath := path.Join(v.p.Path, "issue", v.p.Role)

	reqData := IssueCertificateRequest{
		CommonName: v.p.CN,
		IpSans:     v.p.IpSans,
		AltNames:   v.p.AltNames,
		Ttl:        v.p.TTL,
	}

	//nolint:contextcheck // context is passed in via Logical(ctx). false positive
	secret, err := v.vc.Logical(ctx).Write(fullPath, reqData)
	if err != nil {
		return nil, err
	}

	crtPEM := []byte(secret.Data["certificate"].(string))
	keyPEM := []byte(secret.Data["private_key"].(string))

	cert, err := tls.X509KeyPair(crtPEM, keyPEM)

	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.cachedCertificate = &cert
	if v.p.FileCache.Enabled {
		err := v.CacheCertToFile(&cert)
		if err != nil {
			return &cert, err
		}
	}
	return &cert, err
}

// CacheCertToFile will write out a cert and key to files based on configured path and prefix
func (v *VaultProvider) CacheCertToFile(cert *tls.Certificate) error {
	certfilepath := v.p.FileCache.Path + v.p.FileCache.Prefix + certsource.CertSuffix
	keyfilepath := v.p.FileCache.Path + v.p.FileCache.Prefix + certsource.KeySuffix
	return certsource.CacheCertToFile(cert, certfilepath, keyfilepath)
}

// CacheCaToFile writes the provided ca cert pool to a file based on the provided config
func (v *VaultProvider) CacheCaToFile(pemData []byte) error {
	cafilepath := v.p.FileCache.Path + v.p.FileCache.Prefix + certsource.CaSuffix
	return certsource.CacheCaToFile(pemData, cafilepath)
}

type IssueCertificateRequest struct {
	CommonName string `json:"common_name,omitempty"`
	Ttl        string `json:"ttl,omitempty"`
	AltNames   string `json:"alt_names,omitempty"`
	IpSans     string `json:"ip_sans,omitempty"`
}
