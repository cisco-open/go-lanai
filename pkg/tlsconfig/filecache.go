package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

const (
	CertSuffix = "-cert.pem"
	KeySuffix  = "-key.pem"
	CaSuffix   = "-ca.pem"
)

// CacheCertToFile will write out a cert and key to files based on configured path and prefix
func (v *VaultProvider) CacheCertToFile(cert *tls.Certificate) error {
	if len(cert.Certificate) < 1 {
		return fmt.Errorf("no certificates present in provided tls.Certificate")
	}

	certfilepath := v.p.FileCache.Path + v.ProviderCommon.p.FileCache.Prefix + CertSuffix
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Certificate[0],
	}

	pemBytes := pem.EncodeToMemory(pemBlock)
	if pemBytes == nil {
		return fmt.Errorf("failed to encode certificate to PEM")
	}

	err := os.WriteFile(certfilepath, pemBytes, 0600)
	if err != nil {
		return fmt.Errorf("failed to write PEM data to file: %v", err)
	}

	keyfilepath := v.p.FileCache.Path + v.ProviderCommon.p.FileCache.Prefix + KeySuffix
	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to marshal private key: %v", err)
	}

	privKeyPem := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	}

	pemBytes = pem.EncodeToMemory(privKeyPem)
	if pemBytes == nil {
		return fmt.Errorf("failed to encode private key to PEM")
	}

	err = os.WriteFile(keyfilepath, pemBytes, 0600)
	if err != nil {
		return fmt.Errorf("failed to write private key PEM data to file: %v", err)
	}
	return nil
}

// CacheCaToFile writes the provided ca cert pool to a file based on the provided config
func (v *VaultProvider) CacheCaToFile(pemData []byte) error {
	cafilepath := v.p.FileCache.Path + v.ProviderCommon.p.FileCache.Prefix + CaSuffix

	err := os.WriteFile(cafilepath, pemData, 0600)
	if err != nil {
		return fmt.Errorf("failed to write PEM data to file: %v", err)
	}
	return nil
}
