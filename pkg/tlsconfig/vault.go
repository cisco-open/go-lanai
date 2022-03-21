package tlsconfig

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"io"
	"io/ioutil"
	"path"
	"sync"
	"time"
)

type VaultProvider struct {
	vc *vault.Client
	p Properties

	mutex sync.RWMutex
	cachedCertificate *tls.Certificate
}

func NewVaultProvider(vc *vault.Client, p Properties) *VaultProvider {
	return &VaultProvider{
		vc: vc,
		p: p,
	}
}

func (v *VaultProvider) GetClientCertificate(ctx context.Context) (func(*tls.CertificateRequestInfo) (*tls.Certificate, error), error) {
	err := v.generateClientCertificate(ctx)
	if err != nil {
		return nil, err
	}

	go v.renew(ctx)

	return func(certificate *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			v.mutex.RLock()
			defer v.mutex.RUnlock()
			e := certificate.SupportsCertificate(v.cachedCertificate)
			if e != nil {
				// No acceptable certificate found. Don't send a certificate.
				// see tls package's func (c *Conn) getClientCertificate(cri *CertificateRequestInfo) (*Certificate, error)
				return new(tls.Certificate), nil
			} else {
				return v.cachedCertificate, nil
			}
	}, nil
}

func (v *VaultProvider) RootCAs(ctx context.Context) (*x509.CertPool, error) {
	fullPath := path.Join("/v1/", v.p.Path, "ca", "pem")

	r := v.vc.NewRequest("GET", fullPath)
	resp, err := v.vc.RawRequestWithContext(ctx, r)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	pemBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(pemBytes)

	return certPool, nil
}

func (v *VaultProvider) generateClientCertificate(ctx context.Context) error {
	fullPath := path.Join(v.p.Path, "issue", v.p.Role)

	reqData := IssueCertificateRequest{
		CommonName: v.p.CN,
		IpSans: v.p.IpSans,
		AltNames: v.p.AltNames,
		Ttl: v.p.Ttl,
	}

	secret, err := v.vc.Logical(ctx).Write(fullPath, reqData)
	if err != nil {
		return err
	}

	crtPEM := []byte(secret.Data["certificate"].(string))
	keyPEM := []byte(secret.Data["private_key"].(string))

	cert, err := tls.X509KeyPair(crtPEM, keyPEM)

	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.cachedCertificate = &cert
	return err
}

func (v *VaultProvider) renew(ctx context.Context) {
	for {
		duration := v.nextRenew()
		logger.Infof("certificate will be renewed in %v", duration)
		t := time.NewTimer(duration)

		select {
		case <-ctx.Done():
			if !t.Stop() {
				<-t.C
			}
			break
		case <-t.C:
			//ignore error since we will just schedule another renew
			err := v.generateClientCertificate(ctx)
			if err != nil {
				logger.Warn("certificate renew failed: %v", err)
			} else {
				logger.Infof("certificate has been renewed")
			}
		}
	}
}

// half way between now and cached certificate expiration.
func (v *VaultProvider) nextRenew() time.Duration {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	minDuration := time.Minute
	if v.p.MaxRenewFrequency != ""{
		minDurationConfig, err := time.ParseDuration(v.p.MaxRenewFrequency)
		if err != nil {
			minDuration = minDurationConfig
		}
	}

	parsedCert, err := x509.ParseCertificate(v.cachedCertificate.Certificate[0])
	if err != nil {
		return minDuration
	}

	validTo := parsedCert.NotAfter
	now := time.Now()

	if validTo.Before(now) {
		return minDuration
	}

	durationRemain := validTo.Sub(now)
	next := durationRemain/2
	return next
}

type IssueCertificateRequest struct {
	CommonName string `json:"common_name,omitempty"`
	Ttl        string `json:"ttl,omitempty"`
	AltNames   string `json:"alt_names,omitempty"`
	IpSans     string `json:"ip_sans,omitempty"`
}