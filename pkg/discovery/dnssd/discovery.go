package dnssd

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/utils/loop"
	"regexp"
	"sync"
	"time"
)

type ClientOptions func(opt *ClientConfig)

type ClientConfig struct {
	Logger  log.ContextualLogger
	Verbose bool

	// DNSServerAddr is the address and port of DNS server. e.g. "8.8.8.8:53"
	DNSServerAddr string
	// FQDNTemplate see DiscoveryProperties.FQDNTemplate
	FQDNTemplate string
	// SRVProto see DiscoveryProperties.SRVProto
	SRVProto string
	// SRVService see DiscoveryProperties.SRVService
	SRVService string
	// RefreshInterval interval for background refresher.
	// Note: Foreground refresh happens everytime when Instancer.Service or Instancer.Instances is invoked.
	//       Background refresh is for callbacks only
	// Default: 30s
	RefreshInterval time.Duration
	// FallbackHostMappings known host mappings including default template. Used when DNS lookup fails.
	// See DiscoveryProperties.Fallback.
	// Note: Default mapping should be at the end of the list with `.*` as service name pattern
	FallbackHostMappings []HostMapping
}

type HostMapping struct {
	// ServiceRegex the compiled regular expression HostMappingProperties.Service
	ServiceRegex *regexp.Regexp
	// Hosts is a list of known hosts. Each entry should be a golang template with single-line output.
	// See HostMappingProperties.Hosts
	Hosts []string
}

type dnsDiscoveryClient struct {
	ctx        context.Context
	instancers map[string]*Instancer
	mutex      sync.Mutex
	config     ClientConfig
}

func NewDiscoveryClient(ctx context.Context, opts ...ClientOptions) discovery.Client {
	client := dnsDiscoveryClient{
		ctx:        ctx,
		instancers: map[string]*Instancer{},
		config: ClientConfig{
			Logger:          logger,
			Verbose:         false,
			RefreshInterval: defaultRefreshInterval,
		},
	}

	for _, fn := range opts {
		fn(&client.config)
	}
	return &client
}

func (c *dnsDiscoveryClient) Context() context.Context {
	return c.ctx
}

func (c *dnsDiscoveryClient) Instancer(serviceName string) (discovery.Instancer, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("empty service name")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	instancer, ok := c.instancers[serviceName]
	if ok {
		return instancer, nil
	}
	var e error
	instancer, e = NewInstancer(c.ctx, func(opt *InstancerOption) {
		opt.Name = serviceName
		opt.Logger = c.config.Logger
		opt.Verbose = c.config.Verbose
		opt.DNSServerAddr = c.config.DNSServerAddr
		opt.FQDNTemplate = c.config.FQDNTemplate
		opt.SRVProto = c.config.SRVProto
		opt.SRVService = c.config.SRVService
		opt.HostTemplates = c.findFallbackHostTemplates(serviceName)
		opt.RefresherOptions = []loop.TaskOptions{loop.FixedRepeatInterval(c.config.RefreshInterval)}
	})
	if e == nil {
		c.instancers[serviceName] = instancer
	}
	return instancer, e
}

func (c *dnsDiscoveryClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, v := range c.instancers {
		v.Stop()
	}
	return nil
}

func (c *dnsDiscoveryClient) findFallbackHostTemplates(svcName string) []string {
	for _, mapping := range c.config.FallbackHostMappings {
		if mapping.ServiceRegex.Match([]byte(svcName)) {
			return mapping.Hosts
		}
	}
	return nil
}
