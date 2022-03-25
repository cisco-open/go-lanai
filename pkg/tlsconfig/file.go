package tlsconfig

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
)

type FileProvider struct {
	p Properties
}

func NewFileProvider(p Properties) *FileProvider {
	return &FileProvider{
		p: p,
	}
}

func (f *FileProvider) GetClientCertificate(ctx context.Context) (func(*tls.CertificateRequestInfo) (*tls.Certificate, error), error) {
	return func(certificateReq *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		keyFile, err := os.Open(f.p.KeyFile)
		if err != nil {
			return nil, err
		}

		keyBytes, err := ioutil.ReadAll(keyFile)
		if err != nil {
			return nil, err
		}
		if f.p.KeyPass != "" {
			keyBlock, _ := pem.Decode(keyBytes)
			//nolint:staticcheck
			unEncryptedKey, e := x509.DecryptPEMBlock(keyBlock, []byte("foobar"))
			if e != nil {
			}
			keyBlock.Bytes = unEncryptedKey
			keyBlock.Headers = nil
			keyBytes = pem.EncodeToMemory(keyBlock)
		}
		certfile, err := os.Open(f.p.CertFile)
		if err != nil {
			return nil, err
		}
		certBytes, err := ioutil.ReadAll(certfile)
		if err != nil {
			return nil, err
		}
		clientCert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, err
		}

		e := certificateReq.SupportsCertificate(&clientCert)
		if e != nil {
			// No acceptable certificate found. Don't send a certificate.
			// see tls package's func (c *Conn) getClientCertificate(cri *CertificateRequestInfo) (*Certificate, error)
			return new(tls.Certificate), nil
		} else {
			return &clientCert, nil
		}
	}, nil
}

func (f *FileProvider) RootCAs(ctx context.Context) (*x509.CertPool, error) {
	caPem, err := ioutil.ReadFile(f.p.CaCertFile)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caPem)
	return certPool, nil
}