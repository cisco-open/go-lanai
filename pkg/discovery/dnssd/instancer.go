package dnssd

import (
	"bytes"
	"context"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/discovery/sd"
	"github.com/cisco-open/go-lanai/pkg/utils/loop"
	"net"
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
			RefreshBackoffFactor: sd.DefaultRefreshBackoffFactor,
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

func (i *Instancer) resolveInstancesTask() func(ctx context.Context, _ *loop.Loop) (*discovery.Service, error) {
	return func(ctx context.Context, _ *loop.Loop) (*discovery.Service, error) {
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
	instances := make([]*discovery.Instance, 0)
	for _, srv := range srvs {
		inst := &discovery.Instance{
			ID:      net.JoinHostPort(srv.Target, strconv.Itoa(int(srv.Port))),
			Service: i.Name,
			Address: srv.Target,
			Port:    int(srv.Port),
			Meta: map[string]string{
				kMetaSRVService: i.srvService,
				kMetaSRVProto:   i.srvProto,
				kMetaSRVName:    name,
			},
			Health:   discovery.HealthPassing,
			RawEntry: srv,
		}
		instances = append(instances, inst)
	}
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
