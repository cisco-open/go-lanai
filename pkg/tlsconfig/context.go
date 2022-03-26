package tlsconfig

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"errors"
	"fmt"
	"io"
)

const vaultType = "vault"
const fileType = "file"

type Provider interface {
	io.Closer

	// TODO: VerifyPeerCertificate this can be useful when we need to rotate CAs
	// see https://github.com/golang/go/issues/22836
	// VerifyPeerCertificate() func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error

	// GetClientCertificate this should return a function that returns the client certificate
	GetClientCertificate(ctx context.Context) (func (*tls.CertificateRequestInfo) (*tls.Certificate, error), error)

	// RootCAs this should return the root ca.
	RootCAs(ctx context.Context) (*x509.CertPool, error)
}

type ProviderFactory struct {
	vc *vault.Client
}

func (f *ProviderFactory) GetProvider(properties Properties) (Provider, error) {
	switch properties.Type {
	case vaultType:
		if f.vc != nil {
			return NewVaultProvider(f.vc, properties), nil
		} else {
			return nil, errors.New("can't create vault tls config because there is no vault client")
		}
	case fileType:
		return NewFileProvider(properties), nil
	}
	return nil, errors.New(fmt.Sprintf("%s based tls config provider is not supported", properties.Type))
}