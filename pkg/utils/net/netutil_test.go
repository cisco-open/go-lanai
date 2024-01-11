package netutil

import (
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"net"
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

func TestGetIP(t *testing.T) {
	g := gomega.NewWithT(t)
	ifaces, e := net.Interfaces()
	if e != nil {
		_, e := GetIp("whatever")
		g.Expect(e).To(HaveOccurred(), "there is no network interfaces")
	}
	for _, iface := range ifaces {
		ip, e := GetIp(iface.Name)
		g.Expect(e).To(Succeed(), "GetIp should not fail on %s", "lo")
		g.Expect(ip).To(MatchRegexp(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`), "GetIp(%s) should be correct", "lo")
	}

}