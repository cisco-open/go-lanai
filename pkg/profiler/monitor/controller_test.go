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

package monitor

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/embedded"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
	"time"
)

const (
	kContentType      = `Content-Type`
	ContentTypeBinary = `application/octet-stream`
	ContentTypeHTML   = `text/html`
	ContentTypeText   = `text/plain`
	ContentTypeJS     = `application/javascript`
	ContentTypeJSON   = `application/json`
)

/*************************
	Test
 *************************/

func TestController(t *testing.T) {
	SamplingRate = 10 * time.Millisecond
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		embedded.WithRedis(),
		webtest.WithRealServer(), // Data feed uses websockets, need real server to test
		//apptest.WithTimeout(time.Minute),
		apptest.WithModules(Module, redis.Module),
		test.GomegaSubTest(SubTestStaticAssets(), "TestStaticAssets"),
		test.GomegaSubTest(SubTestChartUI(), "TestChartUI"),
		test.GomegaSubTest(SubTestData(), "TestData"),
		test.GomegaSubTest(SubTestDataFeed(), "TestDataFeed"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestStaticAssets() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		AssertEndpoint(ctx, g, http.MethodGet, "debug/charts/static/main.js", http.StatusOK, ContentTypeJS)
	}
}

func SubTestChartUI() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		AssertEndpoint(ctx, g, http.MethodGet, "debug/charts/", http.StatusOK, ContentTypeHTML)
	}
}

func SubTestData() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		AssertEndpoint(ctx, g, http.MethodGet, "debug/charts/data", http.StatusOK, ContentTypeJSON)
	}
}

func SubTestDataFeed() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		PongTimeout = SamplingRate * 11
		url := fmt.Sprintf("ws://127.0.0.1:%d%s/debug/charts/data-feed", webtest.CurrentPort(ctx), webtest.CurrentContextPath(ctx))

		// Connect to the server
		ws, resp, e := websocket.DefaultDialer.DialContext(ctx, url, nil)
		g.Expect(e).To(Succeed(), "creating connection to data feed should not fail")
		defer func() { _ = ws.Close() }()
		g.Expect(resp.StatusCode).To(Equal(http.StatusSwitchingProtocols), "websocket upgrade response should have correct status code")

		// verify feed
        const msgCount = 10
        for i := 0; i < msgCount; i++ {
        	// try to send messages, messages should be discarded
            e := ws.WriteMessage(websocket.TextMessage, []byte("doesn't matter"))
            g.Expect(e).To(Succeed(), "sending message to data feed shouldn't affect anything")
			// try read feed
			var feed Feed
			e = ws.ReadJSON(&feed)
			g.Expect(e).To(Succeed(), "reading message from data feed should not fail")
			g.Expect(feed).ToNot(BeZero(), "data feed should not be zero valued")
        }

		// validate ping-pong behavior
		// 1. we start to drain feeds (this is required to allow incoming control messages being processed)
		go WSDrainIncomingMessage(ctx, ws)

		// 2. stop responding "ping" starting from the 2nd "ping"
		ws.SetPingHandler(WSPingDenier(ws))

		// 3. observe connection closed from server side.
		closedCh := make(chan int, 1)
		defer close(closedCh)
		ws.SetCloseHandler(func(code int, text string) error {
			closedCh<-code
			return nil
		})
		select {
		case code := <-closedCh:
			g.Expect(code).To(Equal(websocket.CloseNormalClosure), "websocket should be closed normally after client stop responding pings")
		case <-ctx.Done():
			t.Errorf("websocket is not closed by server, context deadline exceeded")
		}
	}
}

/*************************
	Sub Tests
 *************************/

func AssertEndpoint(ctx context.Context, g *gomega.WithT, method, path string, expectedSC int, expectedCT string) {
	req := webtest.NewRequest(ctx, method, path, nil)
	resp := webtest.MustExec(ctx, req).Response
	g.Expect(resp).ToNot(BeNil(), "response should not be nil")
	g.Expect(resp.StatusCode).To(Equal(expectedSC), "[%s %s] should respond correct status code", req.Method, req.URL.Path)
	if expectedSC < 200 || expectedSC > 300 {
		return
	}
	g.Expect(resp.Header.Get(kContentType)).To(HavePrefix(expectedCT), "[%s %s] should respond correct content type", req.Method, req.URL.Path)
}

func WSDrainIncomingMessage(ctx context.Context, ws *websocket.Conn) {
LOOP:
	for e := error(nil); e == nil; _, _, e = ws.NextReader() {
		select {
		case <-ctx.Done():
			break LOOP
		default:
		}
	}
}

// WSPingDenier returns a ping handler that stop responding "ping" starting from the 2nd "ping" observe connection closed from server side.
func WSPingDenier(ws *websocket.Conn) func(appData string) error {
	defaultPingHandler := ws.PingHandler()
	return func(appData string) error {
		defer func() { defaultPingHandler = nil }()
		if defaultPingHandler != nil {
			return defaultPingHandler(appData)
		}
		return nil
	}
}
