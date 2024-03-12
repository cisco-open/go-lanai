package dnssd_test

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/miekg/dns"
	"net"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

var logger = log.New("SD.DNS")

var kCtxDNSServer = struct{}{}

const (
	TestProto   = `_tcp`
	TestService = `_http`
)

type DNSServerOptions func(srv *MockedDNSServer)

// WithDNSServer start a mocked DNS server
func WithDNSServer(opts ...DNSServerOptions) test.Options {
	server := &MockedDNSServer{
		Port: 0,
	}
	for _, fn := range opts {
		fn(server)
	}
	return test.WithOptions(
		test.Setup(func(ctx context.Context, t *testing.T) (context.Context, error) {
			if e := server.Start(ctx); e != nil {
				return ctx, e
			}
			return context.WithValue(ctx, kCtxDNSServer, server), nil
		}),
		apptest.WithDynamicProperties(map[string]apptest.PropertyValuerFunc{
			"cloud.discovery.dns.addr": func(ctx context.Context) interface{} {
				return CurrentMockedDNSAddr(ctx)
			},
		}),
		test.Teardown(func(ctx context.Context, t *testing.T) error {
			return server.Stop(ctx)
		}),
	)
}

func CurrentMockedDNSServer(ctx context.Context) *MockedDNSServer {
	server, _ := ctx.Value(kCtxDNSServer).(*MockedDNSServer)
	return server
}

func CurrentMockedDNSAddr(ctx context.Context) string {
	server := CurrentMockedDNSServer(ctx)
	if server == nil {
		return ""
	}
	return server.currentAddr().String()
}

func doWithMockedDNSServer(ctx context.Context, fn func(server *MockedDNSServer) interface{}) interface{} {
	if server, ok := ctx.Value(kCtxDNSServer).(*MockedDNSServer); ok {
		return fn(server)
	}
	return nil
}

type MockedSRV struct {
	Name     string
	Addr     string
	Port     int
	Priority int
	Weight   int
	Healthy  bool
}

func (srv MockedSRV) QName() string {
	return srv.qName(srv.Name)
}

func (srv MockedSRV) Address() string {
	return AddrToDomain(srv.Addr, srv.QName())
}

func (srv MockedSRV) Hash() string {
	return srv.Addr + ":" + strconv.Itoa(srv.Port)
}

func (srv MockedSRV) qName(name string) string {
	if strings.HasSuffix(name, ".") {
		return name
	}
	return name + "."
}

func (srv MockedSRV) ToRR() (dns.RR, error) {
	str := fmt.Sprintf("%s IN 0 SRV %d %d %d %s", srv.QName(), srv.Priority, srv.Weight, srv.Port, srv.Address())
	return dns.NewRR(str)
}

func NewMockedSRV(svc *MockedService) *MockedSRV {
	fqdn := ServiceFQDN(svc.Name)
	return &MockedSRV{
		Name:     fqdn,
		Addr:     "127.0.0.1",
		Port:     svc.Port,
		Priority: 1,
		Weight:   1,
		Healthy:  svc.Healthy,
	}
}

func ServiceFQDN(serviceName string) string {
	return serviceName + ".test.mock"
}

func AddrToDomain(addr, domain string) string {
	if ip := net.ParseIP(addr); ip != nil {
		return fmt.Sprintf(`%x.%s`, ip, domain)
	}
	return addr
}

type MockedDNSServer struct {
	Port       int
	Server     *dns.Server
	MockedSRVs map[string]map[string]*MockedSRV
}

func (s *MockedDNSServer) RegisterSRV(srv *MockedSRV) {
	s.registerSRV(srv.QName(), srv)
	s.registerSRV(fmt.Sprintf(`%s.%s.%s`, TestService, TestProto, srv.QName()), srv)
}

func (s *MockedDNSServer) registerSRV(key string, srv *MockedSRV) {
	if s.MockedSRVs == nil {
		s.MockedSRVs = map[string]map[string]*MockedSRV{}
	}
	srvs, _ := s.MockedSRVs[key]
	if srvs == nil {
		srvs = map[string]*MockedSRV{}
		s.MockedSRVs[key] = srvs
	}
	srvs[srv.Hash()] = srv
}

func (s *MockedDNSServer) DeregisterSRV(srv *MockedSRV) {
	s.deregisterSRV(srv.QName(), srv)
	s.deregisterSRV(fmt.Sprintf(`%s.%s.%s`, TestService, TestProto, srv.QName()), srv)
}

func (s *MockedDNSServer) deregisterSRV(key string, srv *MockedSRV) {
	if s.MockedSRVs == nil {
		return
	}
	srvs, _ := s.MockedSRVs[key]
	if srvs == nil {
		return
	}
	delete(srvs, srv.Hash())
}

func (s *MockedDNSServer) HandlerFunc() func(rw dns.ResponseWriter, r *dns.Msg) {
	return func(rw dns.ResponseWriter, r *dns.Msg) {
		var resp *dns.Msg
		switch r.Opcode {
		case dns.OpcodeQuery:
			resp = s.handleQuery(r)
		default:
			resp = &dns.Msg{}
		}
		for _, q := range r.Question {
			logger.Debugf(`Question: %v`, &q)
		}
		for _, a := range resp.Answer {
			logger.Debugf(`Answer:   %v`, a)
		}
		if e := rw.WriteMsg(resp); e != nil {
			logger.Debugf("Failed to write DNS message: %v", e)
		}
	}
}

func (s *MockedDNSServer) handleQuery(req *dns.Msg) *dns.Msg {
	resp := &dns.Msg{}
	resp.SetReply(req)
	for _, q := range resp.Question {
		switch q.Qtype {
		case dns.TypeA:
			if answer, e := s.answerA(q); e == nil {
				resp.Answer = append(resp.Answer, answer...)
			}
		case dns.TypeAAAA:
			if answer, e := s.answerAAAA(q); e == nil {
				resp.Answer = append(resp.Answer, answer...)
			}
		case dns.TypeSRV:
			if answer, e := s.answerSRV(q); e == nil {
				resp.Answer = append(resp.Answer, answer...)
			}
		}
	}
	return resp
}

// answerA
// dig @localhost -p XXXX <domain>
func (s *MockedDNSServer) answerA(q dns.Question) ([]dns.RR, error) {
	const total = 2
	answer := make([]dns.RR, 0, total)
	for i := 0; i < total; i++ {
		// always map to localhost
		rr, e := dns.NewRR(fmt.Sprintf("%s IN 0 A %s", q.Name, "127.0.0.1"))
		if e != nil {
			return nil, e
		}
		answer = append(answer, rr)
	}
	return answer, nil
}

// answerAAAA
// dig @localhost -p XXXX <domain>
func (s *MockedDNSServer) answerAAAA(q dns.Question) ([]dns.RR, error) {
	const total = 2
	answer := make([]dns.RR, 0, total)
	for i := 0; i < total; i++ {
		// always map to localhost
		rr, e := dns.NewRR(fmt.Sprintf("%s IN 0 AAAA %s", q.Name, "::1"))
		if e != nil {
			return nil, e
		}
		answer = append(answer, rr)
	}
	return answer, nil
}

// answerSRV
// dig @localhost -p 5353 _api._tcp.<domain> SRV
func (s *MockedDNSServer) answerSRV(q dns.Question) ([]dns.RR, error) {
	srvMap, _ := s.MockedSRVs[q.Name]
	srvs := make([]*MockedSRV, 0, len(srvMap))
	for _, v := range srvMap {
		srvs = append(srvs, v)
	}
	sort.SliceStable(srvs, func(i, j int) bool {
		return srvs[i].Priority < srvs[j].Priority
	})
	answer := make([]dns.RR, 0, len(srvs))
	for _, v := range srvs {
		if !v.Healthy {
			continue
		}
		if rr, e := v.ToRR(); e == nil {
			answer = append(answer, rr)
		}
	}
	return answer, nil
}

func (s *MockedDNSServer) Start(ctx context.Context) error {
	const retry = 3
	dns.HandleFunc(".", s.HandlerFunc())
	var err error
	for i := 0; i < retry; i++ {
		if err = s.start(ctx); err == nil {
			time.Sleep(100 * time.Millisecond)
			logger.WithContext(ctx).Infof("DNS server started at 127.0.0.1:%d", s.Port)
			return nil
		}
	}
	return err
}

func (s *MockedDNSServer) Stop(ctx context.Context) error {
	return s.Server.ShutdownContext(ctx)
}

func (s *MockedDNSServer) start(ctx context.Context) error {
	addrStr := ":0"
	if s.Port > 0x7fff {
		addrStr = fmt.Sprintf(":%d", s.Port)
	}
	// start server
	startCH := make(chan struct{}, 1)
	s.Server = &dns.Server{
		Addr:        addrStr,
		Net:         "udp",
		ReadTimeout: time.Minute,
		ReusePort:   true,
		ReuseAddr:   true,
		NotifyStartedFunc: func() {
			close(startCH)
		},
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	go func() {
		logger.WithContext(ctx).Infof("Starting mocked DNS server at 127.0.0.1%s", addrStr)
		e := s.Server.ListenAndServe()
		logger.WithContext(ctx).Infof("DNS server stopped - Error[%v]", e)
	}()
	// wait for server to start
	shudownFunc := func() { _ = s.Server.Shutdown() }
	select {
	case <-startCH:
	case <-ctx.Done():
		defer shudownFunc()
		return fmt.Errorf(`unable to start mocked DNS server: %v`, ctx.Err())
	}

	addr := s.currentAddr()
	if addr == nil {
		defer shudownFunc()
		return fmt.Errorf(`unable to start mocked DNS server: unknown UDP address`)
	}

	if e := s.TestLookup(ctx); e != nil {
		defer shudownFunc()
		return e
	}
	s.Port = addr.Port
	return nil
}

func (s *MockedDNSServer) TestLookup(ctx context.Context) error {
	addr := s.currentAddr()
	if addr == nil {
		return fmt.Errorf("unknown UDP address")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	// try lookup once to make sure it's started
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			conn, e := d.DialContext(ctx, network, addr.String())
			logger.WithContext(ctx).Debugf("Dialing [%p] %s %s: %T[%v] - Error[%v]", &d, network, addr.String(), conn, conn, e)
			return conn, e
		},
	}

	addrs, e := resolver.LookupNetIP(ctx, "ip4", "anything.test")
	logger.WithContext(ctx).Infof("Test lookup result: IP%v Error[%v]", addrs, e)
	return e
}

func (s *MockedDNSServer) currentAddr() *net.UDPAddr {
	if s.Server.PacketConn == nil {
		return nil
	}
	addr := s.Server.PacketConn.LocalAddr()
	switch v := addr.(type) {
	case *net.UDPAddr:
		return v
	default:
		return nil
	}
}


