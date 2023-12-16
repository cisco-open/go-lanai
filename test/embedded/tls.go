package embedded

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/fs"
	"os"
)

type TLSCerts struct {
	FS   fs.FS
	Cert string
	Key  string
	CAs  []string
}

// ServerTLSWithCerts construct a tls.Config with certificates in a given filesystem.
// The setup Server TLS, following config are required:
// - filesystem to load files from. Default to "."
// - path of certificate file in PEM format, default to "testdata/server.crt"
// - path of certificate private key file in unencrypted PEM format, default to "testdata/server.key"
// - path of at least one CA certificate in PEM format, default to "testdata/ca.crt"
// Note: if any file is missing or not readable, the result tls.Config might not works as expected
func ServerTLSWithCerts(opts ...func(src *TLSCerts)) (*tls.Config, error) {
	src := TLSCerts{
		FS:   os.DirFS("."),
		Cert: "testdata/server.crt",
		Key:  "testdata/server.key",
		CAs:  []string{"testdata/ca.crt"},
	}
	for _, fn := range opts {
		fn(&src)
	}
	// start to load
	caPool := x509.NewCertPool()
	for _, path := range src.CAs {
		pemBytes, e := fs.ReadFile(src.FS, path)
		if e != nil {
			return nil, fmt.Errorf("unable to read CA file [%s]: %v", path, e)
		}
		caPool.AppendCertsFromPEM(pemBytes)
	}
	certBytes, e := fs.ReadFile(src.FS, src.Cert)
	if e != nil {
		return nil, fmt.Errorf("unable to read certificate file [%s]: %v", src.Cert, e)
	}
	keyBytes, e := fs.ReadFile(src.FS, src.Key)
	if e != nil {
		return nil, fmt.Errorf("unable to read private key file [%s]: %v", src.Key, e)
	}
	cert, e := tls.X509KeyPair(certBytes, keyBytes)
	if e != nil {
		return nil, fmt.Errorf("unable to parse certificate: %v", e)
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		RootCAs:      caPool,
		Certificates: []tls.Certificate{cert},
	}, nil
}
