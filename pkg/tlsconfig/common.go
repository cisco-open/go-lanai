package tlsconfig

import "crypto/tls"

var tlsVersions = map[string]uint16{
	"":      tls.VersionTLS10, // default in golang
	"tls10": tls.VersionTLS10,
	"tls11": tls.VersionTLS11,
	"tls12": tls.VersionTLS12,
	"tls13": tls.VersionTLS13,
}

type ProviderCommon struct {
	p Properties
}

func (c ProviderCommon) GetMinTlsVersion() uint16 {
	return tlsVersions[c.p.MinVersion]
}