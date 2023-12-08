package vaultcerts

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	certsource "cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs/source"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/loop"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"fmt"
	"io"
	"path"
	"strings"
	"sync"
	"time"
)

type VaultProvider struct {
	p SourceProperties

	vc    *vault.Client
	cache *certsource.FileCache

	once              sync.Once
	mutex             sync.RWMutex
	cachedCertificate *tls.Certificate

	lcCtx         context.Context
	monitor       *loop.Loop
	monitorCancel context.CancelFunc
}

func NewVaultProvider(ctx context.Context, vc *vault.Client, p SourceProperties) certs.Source {
	if ctx == nil {
		ctx = context.Background()
	}
	cache, e := certsource.NewFileCache(func(opt *certsource.FileCacheOption) {
		opt.Root = p.CachePath
		opt.Type = sourceType
		opt.Prefix = resolveCacheKey(&p)
	})
	if e != nil {
		logger.WithContext(ctx).Warnf("file cache for %s certificate source is not enabled: %v", sourceType, e)
	}
	return &VaultProvider{
		p:       p,
		vc:      vc,
		lcCtx:   ctx,
		cache:   cache,
		monitor: loop.NewLoop(),
	}
}

func (v *VaultProvider) TLSConfig(ctx context.Context, _ ...certs.TLSOptions) (*tls.Config, error) {
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

func (v *VaultProvider) Files(ctx context.Context) (*certs.CertificateFiles, error) {
	if e := v.LazyInit(ctx); e != nil {
		return nil, e
	}
	if v.cache == nil {
		return nil, fmt.Errorf("unable to access certificates as local files: file cache is not enabled for source [%s]", sourceType)
	}

	return &certs.CertificateFiles{
		RootCAPaths:     []string{v.cache.ResolvePath(certsource.CachedFileKeyCA)},
		CertificatePath: v.cache.ResolvePath(certsource.CachedFileKeyCertificate),
		PrivateKeyPath:  v.cache.ResolvePath(certsource.CachedFileKeyPrivateKey),
	}, nil
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
	if v.cache != nil {
		if err := v.cache.CachePEM(pemBytes, certsource.CachedFileKeyCA); err != nil {
			logger.WithContext(ctx).Warnf(`unable to cache CA: %v`, err)
			return certPool, err
		}
	}
	return certPool, nil
}

func (v *VaultProvider) LazyInit(ctx context.Context) error {
	var err error
	v.once.Do(func() {
		// At least get RootCA once
		// TODO should we renew RootCA periodically?
		_, err = v.RootCAs(ctx)
		if err != nil {
			logger.Errorf("Failed to get CAs from Vault: %s", err.Error())
			return
		}
		// At least get Certificate once
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
			logger.Debugf("certificate has been renewed")
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
		TTL:        v.p.TTL,
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
	if v.cache != nil {
		if err := v.cache.CacheCertificate(&cert); err != nil {
			logger.WithContext(ctx).Warnf(`unable to cache certificate: %v`, err)
			return &cert, err
		}
	}
	return &cert, err
}

var (
	cacheKeyReplacer = strings.NewReplacer(
		" ", "-",
		".", "-",
		"_", "-",
		"@", "-at-",
		"/", "-",
		"\\", "-",
	)
	cacheKeyCount = 0
)

func resolveCacheKey(p *SourceProperties) (key string) {
	cacheKeyCount++
	key = fmt.Sprintf(`%s-%s-%d`, p.Role, p.CN, cacheKeyCount)
	key = cacheKeyReplacer.Replace(key)
	return key
}

type IssueCertificateRequest struct {
	CommonName string        `json:"common_name,omitempty"`
	TTL        utils.Duration `json:"ttl,omitempty"`
	AltNames   string        `json:"alt_names,omitempty"`
	IpSans     string        `json:"ip_sans,omitempty"`
}
