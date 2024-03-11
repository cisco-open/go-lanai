package dnssd

import (
	"bytes"
	"context"
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
	defaultRefreshInterval = 1 * time.Minute
	defaultLookupTimeout = 1 * time.Second
)

type InstancerOptions func(opt *InstancerOption)

type InstancerOption struct {
	sd.InstancerOption
	DNSServerAddr     string
	SRVTargetTemplate string
	SRVProto          string
	SRVService        string
}

type Instancer struct {
	sd.CachedInstancer
	resolver   *net.Resolver
	srvTarget  string
	srvProto   string
	srvService string
}

func NewInstancer(ctx context.Context, opts ...InstancerOptions) (*Instancer, error) {
	opt := InstancerOption{
		InstancerOption: sd.InstancerOption{
			Logger: logger,
			RefresherOptions: []loop.TaskOptions{loop.FixedRepeatInterval(defaultRefreshInterval)},
		},
	}
	for _, f := range opts {
		f(&opt)
	}
	var dial func(ctx context.Context, network, address string) (net.Conn, error)
	if len(opt.DNSServerAddr) != 0 {
		addr := opt.DNSServerAddr
		dial = func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		}
	}
	target, e := srvTargetWithTemplate(opt)
	if e != nil {
		return nil, e
	}
	i := &Instancer{
		CachedInstancer: sd.MakeCachedInstancer(func(baseOpt *sd.InstancerOption) {
			*baseOpt = opt.InstancerOption
		}),
		resolver: &net.Resolver{
			Dial:     dial,
		},
		srvTarget:  target,
		srvProto:   strings.TrimLeft(strings.TrimSpace(opt.SRVProto), "_"),
		srvService: strings.TrimLeft(strings.TrimSpace(opt.SRVService), "_"),
	}
	i.RefreshFunc = i.resolveInstancesTask()
	i.Start(ctx)
	return i, nil
}

func (i *Instancer) Service() (svc *discovery.Service) {
	_, _ = i.RefreshNow(context.Background())
	return i.CachedInstancer.Service()
}

func (i *Instancer) Instances(matcher discovery.InstanceMatcher) (ret []*discovery.Instance, err error) {
	_, _ = i.RefreshNow(context.Background())
	return i.CachedInstancer.Instances(matcher)
}

func (i *Instancer) resolveInstancesTask() func(ctx context.Context, _ *loop.Loop) (*discovery.Service, error) {
	return func(ctx context.Context, _ *loop.Loop) (*discovery.Service, error) {
		ctx, cancel := context.WithTimeout(ctx, defaultLookupTimeout)
		defer cancel()
		name, srvs, e := i.resolver.LookupSRV(ctx, i.srvService, i.srvProto, i.srvTarget)
		instances := i.makeInstances(name, srvs)
		svc := &discovery.Service{
			Name:  i.Name,
			Insts: instances,
			Time:  time.Now(),
			Err:   e,
		}
		return svc, e
	}
}

func (i *Instancer) makeInstances(name string, srvs []*net.SRV) []*discovery.Instance {
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

/*******************
	Helpers
 *******************/

type srvTmplData struct {
	ServiceName string
}

func srvTargetWithTemplate(opt InstancerOption) (string, error) {
	tmpl, e := template.New("srv-name").Parse(opt.SRVTargetTemplate)
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

