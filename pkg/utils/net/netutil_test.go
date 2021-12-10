package netutil

import (
	"github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestGetForwardedHostName(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://saml.ciscomsx.com/auth/v2/authorize", nil)
	host := GetForwardedHostName(req)

	g := gomega.NewWithT(t)
	g.Expect(host).To(gomega.Equal("saml.ciscomsx.com"))

	req, _ = http.NewRequest("GET", "https://saml.ciscomsx.com:443/auth/v2/authorize", nil)
	host = GetForwardedHostName(req)
	g.Expect(host).To(gomega.Equal("saml.ciscomsx.com"))

	req, _ = http.NewRequest("GET", "https://192.168.0.1:443/auth/v2/authorize", nil)
	req.Header.Set("X-Forwarded-Host", "saml.ciscomsx.com")
	host = GetForwardedHostName(req)
	g.Expect(host).To(gomega.Equal("saml.ciscomsx.com"))

	req, _ = http.NewRequest("GET", "https://192.168.0.1:443/auth/v2/authorize", nil)
	req.Header.Set("X-Forwarded-Host", "saml.ciscomsx.com:443")
	host = GetForwardedHostName(req)
	g.Expect(host).To(gomega.Equal("saml.ciscomsx.com"))
}