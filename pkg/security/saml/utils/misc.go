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

package samlutils

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/beevik/etree"
	"github.com/russellhaering/goxmldsig/etreeutils"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

// FindChild search direct child XML element matching given NS and Tag in the given parent element
func FindChild(parentEl *etree.Element, childNS string, childTag string) (*etree.Element, error) {
	for _, childEl := range parentEl.ChildElements() {
		if childEl.Tag != childTag {
			continue
		}

		ctx, err := etreeutils.NSBuildParentContext(childEl)
		if err != nil {
			return nil, err
		}
		ctx, err = ctx.SubContext(childEl)
		if err != nil {
			return nil, err
		}

		ns, err := ctx.LookupPrefix(childEl.Space)
		if err != nil {
			return nil, fmt.Errorf("[%s]:%s cannot find prefix %s: %v", childNS, childTag, childEl.Space, err)
		}
		if ns != childNS {
			continue
		}

		return childEl, nil
	}
	return nil, nil
}

// WritePostBindingHTML takes HTML of a request/response submitting form and wrap it in HTML document with proper
// script security tags and send it to given ResponseWriter
func WritePostBindingHTML(formHtml []byte, rw http.ResponseWriter) error {
	body := []byte(fmt.Sprintf(`<!DOCTYPE html><html><body>%s</body></html>`, formHtml))
	csp := fmt.Sprintf("default-src; script-src %s; reflected-xss block; referrer no-referrer;", scriptSrcHash(body))
	rw.Header().Add("Content-Type", "text/html")
	rw.Header().Add("Content-Security-Policy", csp)
	_, e := rw.Write(body)
	return e
}

// scriptSrcHash returns '<hash-algorithm>-<base64-value>' of all inline <script></script> found in given html, delimited by space
// See CSP specs
func scriptSrcHash(htmlBytes []byte) string {
	const fallback = `'unsafe-inline'`
	root, e := html.Parse(bytes.NewReader(htmlBytes))
	if e != nil {
		return fallback
	}
	scripts := findAllHtmlNodes(root, func(node *html.Node) bool {
		return node.Type == html.TextNode && node.Parent != nil && node.Parent.Data == "script"
	})

	srcs := make([]string, len(scripts))
	for i, node := range scripts {
		hash := sha256.Sum256([]byte(node.Data))
		srcs[i] = fmt.Sprintf("'sha256-%s'", base64.StdEncoding.EncodeToString(hash[:]))
	}
	return strings.Join(srcs, " ")
}

func findAllHtmlNodes(node *html.Node, matcher func(*html.Node) bool) (found []*html.Node) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if matcher(child) {
			found = append(found, child)
		}
		if sub := findAllHtmlNodes(child, matcher); len(sub) != 0 {
			found = append(found, sub...)
		}
	}
	return
}

