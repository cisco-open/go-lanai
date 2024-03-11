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

// Package sd, provide base implementation of discovery.Client and discovery.Instancer.
package sd

import (
	"context"
	"errors"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/utils/loop"
	"github.com/go-kit/kit/sd"
	"reflect"
	"sync"
	"time"
)

const (
	DefaultRefreshBackoffFactor float64 = 1.5
)

type InstancerOptions func(opt *InstancerOption)
type InstancerOption struct {
	// Name service name
	Name string
	// Logger logger to use for
	Logger log.Logger
	// Verbose whether to create logs for internal state changes
	Verbose bool
	// RefresherOptions controls how service refresher to retry refreshing attempt in case of failure.
	// Default is loop.ExponentialRepeatIntervalOnError with initial interval 50ms and factor 1.5
	RefresherOptions []loop.TaskOptions
}

// CachedInstancer implements discovery.Instancer and sd.Instancer.
// CachedInstancer provides common implementation of discovery.Instancer with an internal cache and a background goroutine
// to periodically refresh service cache using provided RefreshFunc.
// See discovery.Instancer
// Note: Implementing sd.Instancer is for compatibility reason, using it involves additional Lock locking. Try use instancer's callback capability instead
type CachedInstancer struct {
	InstancerOption
	RefreshFunc func(ctx context.Context, _ *loop.Loop) (*discovery.Service, error)
	readyCond   *sync.Cond
	cacheMtx    sync.RWMutex // RW Lock for cache
	stateMtx    sync.RWMutex // RW Mutex for state, such as start/stop, callback/subscription update
	looper      *loop.Loop
	loopCtx     context.Context
	cancelFunc  context.CancelFunc
	cache       discovery.ServiceCache
	callbacks   map[interface{}]discovery.Callback
	broadcaster *kitBroadcaster
}

// MakeCachedInstancer returns a CachedInstancer that provide basic implementation of discovery.Instancer
// See discovery.Instancer
func MakeCachedInstancer(opts ...InstancerOptions) CachedInstancer {
	opt := InstancerOption{
		RefresherOptions: []loop.TaskOptions{
			loop.ExponentialRepeatIntervalOnError(50*time.Millisecond, DefaultRefreshBackoffFactor),
		},
	}
	for _, f := range opts {
		f(&opt)
	}
	return CachedInstancer{
		InstancerOption: opt,
		looper:          loop.NewLoop(),
		cache:           NewSimpleServiceCache(),
		callbacks:       map[interface{}]discovery.Callback{},
		broadcaster: &kitBroadcaster{
			chs: map[chan<- sd.Event]struct{}{},
		},
	}
}

// ServiceName implements discovery.Instancer
func (i *CachedInstancer) ServiceName() string {
	return i.Name
}

// Service implements discovery.Instancer
func (i *CachedInstancer) Service() (svc *discovery.Service) {
	// read lock only
	i.cacheMtx.RLock()
	defer i.cacheMtx.RUnlock()
	return i.service()
}

// Instances implements discovery.Instancer
func (i *CachedInstancer) Instances(matcher discovery.InstanceMatcher) (ret []*discovery.Instance, err error) {
	// read lock only
	i.cacheMtx.RLock()
	defer i.cacheMtx.RUnlock()

	svc := i.service()
	if errors.Is(i.loopCtx.Err(), context.Canceled) {
		// looper is stopped, we can't trust our cached result anymore
		return []*discovery.Instance{}, discovery.ErrInstancerStopped
	} else if svc.Err != nil {
		err = svc.Err
	}
	ret = svc.Instances(matcher)
	return
}

func (i *CachedInstancer) Start(ctx context.Context) {
	i.stateMtx.Lock()
	defer i.stateMtx.Unlock()

	if i.loopCtx != nil {
		return
	}
	i.readyCond = sync.NewCond(i.cacheMtx.RLocker())
	i.loopCtx, i.cancelFunc = i.looper.Run(ctx)
	i.looper.Repeat(i.refreshTask(), i.RefresherOptions...)
}

func (i *CachedInstancer) RegisterCallback(id interface{}, cb discovery.Callback) {
	if id == nil || cb == nil {
		return
	}

	i.stateMtx.Lock()
	i.callbacks[id] = cb
	i.stateMtx.Unlock()
	//cb(i)
}

func (i *CachedInstancer) DeregisterCallback(id interface{}) {
	if id == nil {
		return
	}

	i.stateMtx.Lock()
	defer i.stateMtx.Unlock()
	delete(i.callbacks, id)
}

// Stop implements sd.Instancer and CachedInstancer.
func (i *CachedInstancer) Stop() {
	i.stateMtx.Lock()
	defer i.stateMtx.Unlock()
	if i.cancelFunc != nil {
		i.cancelFunc()
	}
}

// Register implements sd.Instancer.
func (i *CachedInstancer) Register(ch chan<- sd.Event) {
	i.stateMtx.Lock()
	i.cacheMtx.RLock()
	defer i.stateMtx.Unlock()
	defer i.cacheMtx.RUnlock()

	var event sd.Event
	svc := i.cache.Get(i.Name)
	if svc != nil {
		event = makeEvent(svc)
	}
	i.broadcaster.register(ch, event)
}

// Deregister implements sd.Instancer.
func (i *CachedInstancer) Deregister(ch chan<- sd.Event) {
	i.stateMtx.Lock()
	defer i.stateMtx.Unlock()

	i.broadcaster.deregister(ch)
}

// RefreshNow invoke refresh task immediately in current goroutine.
// Note: refresh function is run in refresher's loop, and may block depending on RefreshFunc
func (i *CachedInstancer) RefreshNow(ctx context.Context) (*discovery.Service, error) {
	v, e := i.refreshTask()(ctx, i.looper)
	if e != nil {
		return nil, e
	}
	return v.(*discovery.Service), nil
}

// service is not goroutine-safe and returns non-nil *Service.
// It would wait until first RefreshFunc finished and *Service become available
func (i *CachedInstancer) service() (svc *discovery.Service) {
	for !i.cache.Has(i.Name) {
		i.readyCond.Wait()
	}
	return i.cache.Get(i.Name)
}

func (i *CachedInstancer) refreshTask() loop.TaskFunc {
	return func(ctx context.Context, l *loop.Loop) (ret interface{}, err error) {
		service, e := i.RefreshFunc(ctx, l)
		i.onRefresh(ctx, service, e)
		return service, e
	}
}

func (i *CachedInstancer) onRefresh(_ context.Context, service *discovery.Service, err error) {
	i.cacheMtx.Lock()
	var notify bool
	defer func() {
		// we need to release the write lock before invoking callbacks
		i.cacheMtx.Unlock()
		i.readyCond.Broadcast()
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
func (i *CachedInstancer) invokeCallbacks() {
	i.stateMtx.RLock()
	defer i.stateMtx.RUnlock()
	for _, cb := range i.callbacks {
		cb(i)
	}
}

func (i *CachedInstancer) determineFirstErrTime(err error, old *discovery.Service) time.Time {
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

func (i *CachedInstancer) shouldNotify(new, old *discovery.Service) bool {
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
	return !reflect.DeepEqual(new.Insts, old.Insts) ||
		new.Err != nil && old.Err == nil ||
		new.Err == nil && old.Err != nil
}

func (i *CachedInstancer) logUpdate(new, old *discovery.Service) {
	if i.Verbose {
		i.verboseLog(new, old)
	}

	// for regular log, we only log if healthy service changes between 0 and non-zero
	var before, now int
	if old != nil {
		before = old.InstanceCount(discovery.InstanceIsHealthy())
	}
	if new != nil {
		now = new.InstanceCount(discovery.InstanceIsHealthy())
	}
	if before == 0 && now > 0 {
		i.Logger.Infof("service [%s] became available", i.Name)
	} else if before > 0 && now == 0 {
		i.Logger.Warnf("service [%s] healthy instances dropped to 0", i.Name)
	}
}

func (i *CachedInstancer) verboseLog(new, old *discovery.Service) {
	// verbose
	if new != nil && new.Err != nil && (old == nil || old.Err == nil) {
		i.Logger.Infof("error when finding instances for service %s: %v", i.Name, new.Err)
	} else {
		diff := diff(new, old)
		i.Logger.Debugf(`refreshed instances %s: [healthy=%d] [unchanged=%d] [updated=%d] [new=%d] [removed=%d]`, i.Name,
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

func makeEvent(svc *discovery.Service) sd.Event {
	instances := make([]string, len(svc.Insts))
	for i, inst := range svc.Insts {
		instances[i] = fmt.Sprintf("%s:%d", inst.Address, inst.Port)
	}
	return sd.Event{
		Instances: instances,
		Err:       svc.Err,
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
	deleted []*discovery.Instance
}

func diff(new, old *discovery.Service) (ret *svcDiff) {
	ret = &svcDiff{}
	switch {
	case new == nil && old != nil:
		ret.deleted = old.Insts
		return
	case new != nil && old == nil:
		ret.added = new.Insts
		for _, inst := range ret.added {
			if inst.Health == discovery.HealthPassing {
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
		if inst.Health == discovery.HealthPassing {
			ret.healthy = append(ret.healthy, inst)
		}
	}
	return
}
