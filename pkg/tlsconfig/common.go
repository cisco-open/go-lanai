package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/loop"
	"errors"
	"fmt"
	"time"
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

// half way between now and cached certificate expiration.
func (v *ProviderCommon) tryRenewRepeatIntervalFunc() loop.RepeatIntervalFunc {
	return func(result interface{}, err error) (ret time.Duration) {
		defer func() {
			logger.Infof("certificate will renew in %v", ret)
		}()

		minDuration := 1 * time.Minute
		if v.p.MinRenewInterval != "" {
			minDurationConfig, e := time.ParseDuration(v.p.MinRenewInterval)
			if e == nil {
				minDuration = minDurationConfig
			}
		}

		if err != nil {
			return minDuration
		}

		cert := result.(*tls.Certificate)
		if len(cert.Certificate) < 1 {
			return minDuration
		}

		parsedCert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return minDuration
		}

		validTo := parsedCert.NotAfter
		now := time.Now()

		if validTo.Before(now) {
			return minDuration
		}

		durationRemain := validTo.Sub(now)
		next := durationRemain / 2

		if minDuration > next {
			next = minDuration
		}
		return next
	}
}
