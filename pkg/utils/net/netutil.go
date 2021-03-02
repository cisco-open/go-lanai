package netutil

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

//TODO: review - code borrowed from go-msx
func GetIp(iface string) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	var ip net.IP
	for _, i := range ifaces {
		name := i.Name
		if iface != name && strings.Contains(name, "utun") {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil {
					ip = v.IP
				}
			case *net.IPAddr:
				if v.IP.To4() != nil {
					ip = v.IP
				}
			}
		}
		if iface == name {
			break
		}
	}
	if ip == nil {
		if iface == "" {
			return "", errors.New("No valid interface or address found")
		} else {
			return "", errors.New(fmt.Sprintf("Interface %s not found or no address", iface))
		}
	}

	return ip.String(), nil
}
