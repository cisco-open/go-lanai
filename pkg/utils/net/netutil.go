// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package netutil

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

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
		addrs, e := i.Addrs()
		if e != nil {
			return "", e
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			//SuppressWarnings go:S1871 type switching, not duplicate
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

func GetForwardedHostName(request *http.Request) string {
	var host string
	fwdAddress := request.Header.Get("X-Forwarded-Host") // capitalisation doesn't matter
	if fwdAddress != "" {
		ips := strings.Split(fwdAddress, ",")
		orig := strings.TrimSpace(ips[0])
		reqHost, _, err := net.SplitHostPort(orig)
		if err == nil {
			host = reqHost
		} else {
			host = orig
		}
	} else {
		reqHost, _, err := net.SplitHostPort(request.Host)
		if err == nil {
			host = reqHost
		} else {
			host = request.Host
		}
	}
	return host
}

func AppendRedirectUrl(redirectUrl string, params map[string]string) (string, error) {
	loc, e := url.ParseRequestURI(redirectUrl)
	if e != nil || !loc.IsAbs() {
		return "", errors.New("invalid redirect_uri")
	}

	// TODO support fragments
	query := loc.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	loc.RawQuery = query.Encode()

	return loc.String(), nil
}
