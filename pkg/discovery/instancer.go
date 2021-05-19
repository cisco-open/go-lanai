package discovery

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/loop"
	"fmt"
	"github.com/go-kit/kit/sd"
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"reflect"
	"sort"
	"sync"
	"time"
)

const (
	defaultIndex       uint64  = 0
	queryBackoffFactor float64 = 1.5
)

type ConsulInstancerOptions func(opt *ConsulInstancerOption)
type ConsulInstancerOption struct {
	ServiceName      string
	ConsulConnection *consul.Connection
	Logger           log.Logger
	Selector         InstanceMatcher
	Verbose          bool
}

// ConsulInstancer implements sd.Instancer and Instancer.
// It yields service for a serviceName in Consul.
// Note: implementing sd.Instancer is for compatibility reason, using it involves addtional Mutex locking.
// 		 Try use Instancer's callback capability instead
type ConsulInstancer struct {
	cacheMtx    sync.RWMutex // RW Mutex for cache
	stateMtx    sync.RWMutex // RW Mutext for state, such as start/stop, callback/subscription update
	client      kitconsul.Client
	serviceName string
	selector    InstanceMatcher
	looper      *loop.Loop
	loopCtx     context.Context
	cancelFunc  context.CancelFunc
	lastMeta    *api.QueryMeta
	cache       ServiceCache
	callbacks   map[interface{}]Callback
	broadcaster *kitBroadcaster
	logger      log.Logger
	verbose     bool
}

// NewConsulInstancer returns a customized Consul instancer that publishes service for the
// requested serviceName. It only returns service for which the passed tags are present.
//func NewConsulInstancer(ctx context.Context, client kitconsul.Client, logger log.Logger, service string, tags []string, passingOnly bool) *ConsulInstancer {
func NewConsulInstancer(ctx context.Context, opts ...ConsulInstancerOptions) *ConsulInstancer {
	opt := ConsulInstancerOption{}
	for _, f := range opts {
		f(&opt)
	}
	i := &ConsulInstancer{
		client:      kitconsul.NewClient(opt.ConsulConnection.Client()),
		serviceName: opt.ServiceName,
		selector:    opt.Selector,
		logger:      opt.Logger,
		verbose:     opt.Verbose,
		looper:      loop.NewLoop(),
		cache:       newSimpleServiceCache(),
		callbacks:   map[interface{}]Callback{},
		broadcaster: &kitBroadcaster{
			chs: map[chan<- sd.Event]struct{}{},
		},
	}
	i.Start(ctx)
	return i
}

// ServiceName implements Instancer
func (i *ConsulInstancer) ServiceName() string {
	return i.serviceName
}

// Service implements Instancer
func (i *ConsulInstancer) Service() *Service {
	// read lock only
	i.cacheMtx.RLock()
	defer i.cacheMtx.RUnlock()
	return i.cache.Get(i.serviceName)
}

// Instances implements Instancer
func (i *ConsulInstancer) Instances(matcher InstanceMatcher) (ret []*Instance, err error) {
	// read lock only
	i.cacheMtx.RLock()
	defer i.cacheMtx.RUnlock()

	svc := i.cache.Get(i.serviceName)
	if i.loopCtx.Err() == context.Canceled {
		// looper is stopped, we can't trust our cached result anymore
		return []*Instance{}, ErrInstancerStopped
	} else if svc == nil {
		return []*Instance{}, fmt.Errorf("cannot find service [%s]", i.serviceName)
	} else if svc.Err != nil {
		err = svc.Err
	}
	ret = svc.Instances(matcher)
	return
}

func (i *ConsulInstancer) Start(ctx context.Context) {
	i.stateMtx.Lock()
	defer i.stateMtx.Unlock()

	if i.loopCtx != nil {
		return
	}
	i.loopCtx, i.cancelFunc = i.looper.Run(ctx)
	i.looper.Repeat(
		i.resolveInstancesTask(),
		loop.ExponentialRepeatIntervalOnError(50*time.Millisecond, queryBackoffFactor),
	)
}

func (i *ConsulInstancer) RegisterCallback(id interface{}, cb Callback) {
	if id == nil || cb == nil {
		return
	}

	i.stateMtx.Lock()
	i.callbacks[id] = cb
	i.stateMtx.Unlock()
	cb(i)
}

func (i *ConsulInstancer) DeregisterCallback(id interface{}) {
	if id == nil {
		return
	}

	i.stateMtx.Lock()
	defer i.stateMtx.Unlock()
	delete(i.callbacks, id)
}

// Stop implements sd.Instancer and Instancer.
func (i *ConsulInstancer) Stop() {
	i.stateMtx.Lock()
	defer i.stateMtx.Unlock()
	if i.cancelFunc != nil {
		i.cancelFunc()
	}
}

// Register implements sd.Instancer.
func (i *ConsulInstancer) Register(ch chan<- sd.Event) {
	i.stateMtx.Lock()
	i.cacheMtx.RLock()
	defer i.stateMtx.Unlock()
	defer i.cacheMtx.RUnlock()

	var event sd.Event
	svc := i.cache.Get(i.serviceName)
	if svc != nil {
		event = makeEvent(svc)
	}
	i.broadcaster.register(ch, event)
}

// Deregister implements sd.Instancer.
func (i *ConsulInstancer) Deregister(ch chan<- sd.Event) {
	i.stateMtx.Lock()
	defer i.stateMtx.Unlock()

	i.broadcaster.deregister(ch)
}

func (i *ConsulInstancer) resolveInstancesTask() loop.TaskFunc {
	// Note:
	// 		Consul doesn't support more than one tag in its serviceName query method.
	// 		https://github.com/hashicorp/consul/issues/294
	// 		Hashi suggest prepared queries, but they don't support blocking.
	// 		https://www.consul.io/docs/agent/http/query.html#execute
	// 		If we want blocking for efficiency, we can use single tag
	return func(ctx context.Context, _ *loop.Loop) (ret interface{}, err error) {
		// Note: i.lastMeta is only updated in this function, and this function is executed via loop.Loop.
		// 		 because loop.Loop guarantees that all tasks are executed one-by-one,
		// 		 there is no need to use Mutex or locking
		lastIndex := defaultIndex
		if i.lastMeta != nil {
			lastIndex = i.lastMeta.LastIndex
		}

		entries, meta, e := i.client.Service(i.serviceName, "", false, &api.QueryOptions{
			WaitIndex: lastIndex,
		})

		i.lastMeta = meta
		i.processResolvedServiceEntries(ctx, entries, e)
		return nil, e
	}
}

func (i *ConsulInstancer) processResolvedServiceEntries(_ context.Context, entries []*api.ServiceEntry, err error) {
	insts := makeInstances(entries, i.selector)
	service := &Service{
		Name:  i.serviceName,
		Insts: insts,
		Time:  time.Now(),
		Err:   err,
	}

	i.cacheMtx.Lock()
	var notify bool
	defer func() {
		// we need to release the write lock before invoking callbacks
		i.cacheMtx.Unlock()
		if notify {
			i.invokeCallbacks()
		}
	}()

	// record result
	existing := i.cache.Set(service.Name, service)
	service.FirstErrAt = i.determineFirstErrTime(err, existing)

	notify = i.shouldNotify(service, existing)
	if notify {
		i.logUpdate(service, existing)
		// for go-kit compatibility
		evt := makeEvent(service)
		i.broadcaster.broadcast(evt)
	}
}

// invokeCallbacks uses read lock
func (i *ConsulInstancer) invokeCallbacks() {
	i.stateMtx.RLock()
	defer i.stateMtx.RUnlock()
	for _, cb := range i.callbacks {
		cb(i)
	}
}

func (i *ConsulInstancer) determineFirstErrTime(err error, old *Service) time.Time {
	switch {
	case err == nil:
		// happy path, there is no new error, zero time
		return time.Time{}
	case old == nil || old.Err == nil:
		// old record had no error, the err is the first err
		return time.Now()
	default:
		// old record had error, carry over the old error time
		return old.FirstErrAt
	}
}

func (i *ConsulInstancer) shouldNotify(new, old *Service) bool {
	switch {
	case old == nil && new == nil:
		return false
	case old == nil || new == nil:
		return true
	}
	// notify with 3 conditions:
	// 1. service instances changed
	// 2. new service have error and old doesn't
	// 3. old service have error but new doesn't
	return !reflect.DeepEqual(new.Instances, old.Instances) ||
		new.Err != nil && old.Err == nil ||
		new.Err == nil && old.Err != nil
}

func (i *ConsulInstancer) logUpdate(new, old *Service) {
	if !i.verbose {
		return
	}

	if new.Err != nil && old.Err == nil {
		i.logger.Infof("error when finding instances for service %s: %v", i.serviceName, new.Err)
	} else {
		diff := diff(new, old)
		i.logger.Debugf(`refreshed instances %s[healthy=%d]: =%d !%d +%d -%d`, i.serviceName,
			len(diff.healthy), len(diff.unchanged), len(diff.updated), len(diff.added), len(diff.deleted))
	}
}

/***********************
	go-kit event
 ***********************/
// kitBroadcaster is not goroutine safe
type kitBroadcaster struct {
	chs map[chan<- sd.Event]struct{}
}

func (b *kitBroadcaster) broadcast(event sd.Event) {
	for c := range b.chs {
		b.send(event, c)
	}
}

func (b *kitBroadcaster) send(event sd.Event, ch chan<- sd.Event) {
	eventCopy := copyEvent(event)
	ch <- eventCopy
}

func (b *kitBroadcaster) register(c chan<- sd.Event, lastEvent sd.Event) {
	b.chs[c] = struct{}{}
	b.send(lastEvent, c)
}

func (b *kitBroadcaster) deregister(c chan<- sd.Event) {
	delete(b.chs, c)
}

/***********************
	Helpers
 ***********************/
func makeInstances(entries []*api.ServiceEntry, selector InstanceMatcher) []*Instance {
	instances := make([]*Instance, 0)
	for _, entry := range entries {
		addr := entry.Service.Address
		if addr == "" {
			addr = entry.Node.Address
		}
		inst := &Instance{
			ID:       entry.Service.ID,
			Service:  entry.Service.Service,
			Address:  addr,
			Port:     entry.Service.Port,
			Tags:     entry.Service.Tags,
			Meta:     entry.Service.Meta,
			Health:   parseHealth(entry),
			RawEntry: entry,
		}

		if selector == nil {
			instances = append(instances, inst)
		} else if matched, e := selector.Matches(inst); e == nil && matched {
			instances = append(instances, inst)
		}
	}
	sort.SliceStable(instances, func(i, j int) bool {
		return instances[i].ID < instances[j].ID
	})
	return instances
}

func makeEvent(svc *Service) sd.Event {
	instances := make([]string, len(svc.Insts))
	for i, inst := range svc.Insts {
		instances[i] = fmt.Sprintf("%s:%d", inst.Address, inst.Port)
	}
	return sd.Event{
		Instances: instances,
		Err:       svc.Err,
	}
}

func parseHealth(entry *api.ServiceEntry) HealthStatus {
	switch status := entry.Checks.AggregatedStatus(); status {
	case api.HealthPassing:
		return HealthPassing
	case api.HealthWarning:
		return HealthWarning
	case api.HealthCritical:
		return HealthCritical
	case api.HealthMaint:
		return HealthMaintenance
	default:
		return HealthAny
	}
}

// copyEvent does a deep copy on sd.Event
func copyEvent(e sd.Event) sd.Event {
	// observers all need their own copy of event
	// because they can directly modify event.Instances
	// for example, by calling sort.Strings
	if e.Instances == nil {
		return e
	}
	instances := make([]string, len(e.Instances))
	copy(instances, e.Instances)
	e.Instances = instances
	return e
}

type svcDiff struct {
	healthy,
	unchanged,
	updated,
	added,
	deleted []*Instance
}

func diff(new, old *Service) (ret *svcDiff) {
	ret = &svcDiff{}
	switch {
	case new == nil && old != nil:
		ret.deleted = old.Insts
		return
	case new != nil && old == nil:
		ret.added = new.Insts
		for _, inst := range ret.added {
			if inst.Health == HealthPassing {
				ret.healthy = append(ret.healthy, inst)
			}
		}
		return
	case new == nil || old == nil:
		return
	}

	// find differences, Note that we know instances are sorted by ID
	newN, oldN := len(new.Insts), len(old.Insts)
	for newI, oldI := 0, 0; newI < newN && oldI < oldN; {
		newInst, oldInst := new.Insts[newI], old.Insts[oldI]
		switch {
		case newInst.ID > oldInst.ID:
			oldI++
			ret.deleted = append(ret.deleted, oldInst)
		case newInst.ID < oldInst.ID:
			newI++
			ret.added = append(ret.added, newInst)
		default:
			newI++
			oldI++
			if !reflect.DeepEqual(newInst, oldInst) {
				ret.updated = append(ret.updated, newInst)
			} else {
				ret.unchanged = append(ret.unchanged, newInst)
			}
		}
	}

	for _, inst := range new.Insts {
		if inst.Health == HealthPassing {
			ret.healthy = append(ret.healthy, inst)
		}
	}
	return
}
