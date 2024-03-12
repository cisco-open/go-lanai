package dnssd

import (
	"bytes"
	"context"
	"errors"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/discovery/sd"
	"github.com/cisco-open/go-lanai/pkg/utils/loop"
	"net"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	kMetaSRVName    = "_srv_name"
	kMetaSRVService = "_srv_service"
	kMetaSRVProto   = "_srv_proto"
)

var (
	defaultRefreshInterval = 30 * time.Second
	defaultLookupTimeout   = 2 * time.Second
)

type InstancerOptions func(opt *InstancerOption)

type InstancerOption struct {
	sd.InstancerOption
	DNSServerAddr string
	FQDNTemplate  string
	SRVProto      string
	SRVService    string
	FQDNFallback  bool
}

type Instancer struct {
	sd.CachedInstancer
	context      context.Context
	resolver     *net.Resolver
	fqdn         string
	srvProto     string
	srvService   string
	fqdnFallback bool
}

func NewInstancer(ctx context.Context, opts ...InstancerOptions) (*Instancer, error) {
	opt := InstancerOption{
		InstancerOption: sd.InstancerOption{
			Logger:           logger,
			RefresherOptions: []loop.TaskOptions{loop.FixedRepeatInterval(defaultRefreshInterval)},
		},
	}
	for _, f := range opts {
		f(&opt)
	}

	var dial func(ctx context.Context, network, address string) (net.Conn, error)
	if len(opt.DNSServerAddr) != 0 {
		dial = dialWithAddrOverride(opt.DNSServerAddr)
	}

	target, e := fqdnWithTemplate(opt)
	if e != nil {
		return nil, e
	}

	i := &Instancer{
		CachedInstancer: sd.MakeCachedInstancer(func(baseOpt *sd.CachedInstancerOption) {
			baseOpt.InstancerOption = opt.InstancerOption
		}),
		context: ctx,
		resolver: &net.Resolver{
			PreferGo: dial != nil,
			Dial:     dial,
		},
		fqdn:         target,
		srvProto:     strings.TrimLeft(strings.TrimSpace(opt.SRVProto), "_"),
		srvService:   strings.TrimLeft(strings.TrimSpace(opt.SRVService), "_"),
		fqdnFallback: opt.FQDNFallback,
	}
	i.BackgroundRefreshFunc = i.resolveInstancesTask()
	i.ForegroundRefreshFunc = i.resolveInstancesTask()
	i.Start(ctx)
	return i, nil
}

func (i *Instancer) Service() (svc *discovery.Service) {
	_, _ = i.RefreshNow(i.context)
	return i.CachedInstancer.Service()
}

func (i *Instancer) Instances(matcher discovery.InstanceMatcher) (ret []*discovery.Instance, err error) {
	_, _ = i.RefreshNow(i.context)
	return i.CachedInstancer.Instances(matcher)
}

func (i *Instancer) resolveInstancesTask() func(ctx context.Context) (*discovery.Service, error) {
	return func(ctx context.Context) (*discovery.Service, error) {
		instances, e := i.trySRVRecord(ctx)
		if (e != nil || len(instances) == 0) && i.fqdnFallback {
			instances = i.makeInstancesFromAddrs([]string{i.fqdn})
			e = nil
		}
		svc := &discovery.Service{
			Name:  i.Name,
			Insts: instances,
			Time:  time.Now(),
			Err:   e,
		}
		return svc, e
	}
}

func (i *Instancer) trySRVRecord(ctx context.Context) ([]*discovery.Instance, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultLookupTimeout)
	defer cancel()
	name, srvs, e := i.resolver.LookupSRV(ctx, i.srvService, i.srvProto, i.fqdn)
	e = i.translateLookupError(ctx, e)
	if e != nil {
		return nil, e
	}
	return i.makeInstancesFromSRVs(name, srvs), nil
}

func (i *Instancer) makeInstancesFromSRVs(name string, srvs []*net.SRV) []*discovery.Instance {
	instances := make([]*discovery.Instance, len(srvs))
	for j := range srvs {
		instances[j] = &discovery.Instance{
			ID:      net.JoinHostPort(srvs[j].Target, strconv.Itoa(int(srvs[j].Port))),
			Service: i.Name,
			Address: srvs[j].Target,
			Port:    int(srvs[j].Port),
			Meta: map[string]string{
				kMetaSRVService: i.srvService,
				kMetaSRVProto:   i.srvProto,
				kMetaSRVName:    name,
			},
			Health:   discovery.HealthPassing,
			RawEntry: *srvs[j],
		}
	}
	sort.SliceStable(instances, func(i, j int) bool {
		return instances[i].ID < instances[j].ID
	})
	return instances
}

func (i *Instancer) makeInstancesFromAddrs(addrs []string) []*discovery.Instance {
	instances := make([]*discovery.Instance, len(addrs))
	for j := range addrs {
		instances[j] = &discovery.Instance{
			ID:       net.JoinHostPort(addrs[j], "0"),
			Service:  i.Name,
			Address:  addrs[j],
			Meta:     map[string]string{},
			Health:   discovery.HealthPassing,
			RawEntry: addrs[j],
		}
	}
	sort.SliceStable(instances, func(i, j int) bool {
		return instances[i].ID < instances[j].ID
	})
	return instances
}

func (i *Instancer) translateLookupError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.IsNotFound {
			return nil
		}
	}
	i.logError(ctx, err)
	return err
}

func (i *Instancer) logError(ctx context.Context, err error) {
	if i.Verbose {
		i.Logger.WithContext(ctx).Debugf(`failed to lookup %s %s %s IN SRV: %v`, i.srvService, i.srvProto, i.fqdn, err)
	}
}

/*******************
	Helpers
 *******************/

func dialWithAddrOverride(addr string) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, _ string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, network, addr)
	}
}

type srvTmplData struct {
	ServiceName string
}

func fqdnWithTemplate(opt InstancerOption) (string, error) {
	tmpl, e := template.New("srv-name").Parse(opt.FQDNTemplate)
	if e != nil {
		return "", e
	}
	var buf bytes.Buffer
	data := srvTmplData{
		ServiceName: opt.Name,
	}
	if e := tmpl.Execute(&buf, data); e != nil {
		return "", e
	}
	return buf.String(), nil
}
