package tlsconfig

import (
	"crypto/tls"
	"errors"
	"fmt"
)

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

func (c ProviderCommon) GetMinTlsVersion() (uint16, error) {
	if v, ok := tlsVersions[c.p.MinVersion]; ok {
		return v, nil
	} else {
		return tls.VersionTLS10, errors.New(fmt.Sprintf("unsupported min tls version %s", c.p.MinVersion))
	}
}
