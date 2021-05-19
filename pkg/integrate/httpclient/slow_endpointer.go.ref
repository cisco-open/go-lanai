package httpclient

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"io"
	"sync"
	"time"
)

// EndpointerOptions allows control of endpointCache behavior.
type EndpointerOptions func(*EndpointerOption)

type EndpointerOption struct {
	ServiceName       string
	EndpointFactory   EndpointFactory
	Selector          discovery.InstanceMatcher
	InvalidateOnError bool
	InvalidateTimeout time.Duration
	Logger            log.Logger
}

type endpointEntry struct {
	ep endpoint.Endpoint
	closer io.Closer
}

// KitEndpointer implements sd.Endpointer interface and works with discovery.Instancer.
// When created with NewKitEndpointer function, it automatically registers
// as a subscriber to events from the Instances and maintains a list
// of active Endpoints.
type KitEndpointer struct {
	mtx       sync.RWMutex
	instancer discovery.Instancer
	selector  discovery.InstanceMatcher
	factory   EndpointFactory
	entries   map[string]endpointEntry
	endpoints []endpoint.Endpoint
	lastErr   error
	timeout   time.Duration
	logger    log.Logger
}

// NewKitEndpointer creates a custom sd.Endpointer that work with discovery.Instancer
// and uses factory f to create Endpoints. If src notifies of an error, the Endpointer
// keeps returning previously created Endpoints assuming they are still good, unless
// this behavior is disabled via InvalidateOnError option.
func NewKitEndpointer(discClient discovery.Client, opts ...EndpointerOptions) (*KitEndpointer, error) {
	opt := EndpointerOption{
		Selector: matcher.Any(),
		Logger: logger,
	}
	for _, f := range opts {
		f(&opt)
	}

	// some validation
	if opt.ServiceName == "" || opt.EndpointFactory == nil {
		return nil, fmt.Errorf("service name and endpoint factory are required options for KitEndpointer")
	}

	if !opt.InvalidateOnError {
		opt.InvalidateTimeout = -1
	} else if opt.InvalidateTimeout < 0 {
		opt.InvalidateTimeout = 0 // invalidate immediately
	}

	instancer, e := discClient.Instancer(opt.ServiceName)
	if e != nil {
		return nil, e
	}

	// create and start
	ret := &KitEndpointer{
		instancer: instancer,
		selector:  opt.Selector,
		factory:   opt.EndpointFactory,
		logger:    opt.Logger,
		timeout:   opt.InvalidateTimeout,
		entries:   map[string]endpointEntry{},
		endpoints: []endpoint.Endpoint{},
	}
	ret.start()
	return ret, nil
}

// Endpoints implements sd.Endpointer.
func (ke *KitEndpointer) Endpoints() ([]endpoint.Endpoint, error) {
	ke.mtx.RLock()
	defer ke.mtx.RUnlock()
	return ke.endpoints, ke.lastErr
}

// Close deregister itself from discovery.Instancer
func (ke *KitEndpointer) Close() {
	ke.instancer.DeregisterCallback(ke)
	for _, entry := range ke.entries {
		_ = entry.closer.Close()
	}
}

func (ke *KitEndpointer) start() {
	ke.instancer.RegisterCallback(ke, ke.discoveryCallback)
}

// cleanupIfHasError is goroutine-safe
func (ke *KitEndpointer) cleanupIfHasError() {
	if ke.lastErr == nil {
		return
	}
	ke.mtx.Lock()
	defer ke.mtx.Unlock()

	// double check after locking
	if ke.lastErr == nil {
		return
	}
	ke.cleanup()
}

// cleanup is NOT goroutine-safe
func (ke *KitEndpointer) cleanup() {
	ke.entries = map[string]endpointEntry{}
	ke.endpoints = []endpoint.Endpoint{}
}

// handleError is NOT goroutine-safe
func (ke *KitEndpointer) handleError(err error) {
	ke.lastErr = err
	switch {
	case ke.timeout < 0:
		// nothing to do
	case ke.timeout == 0:
		// cleanup now
		ke.cleanup()
	case ke.timeout > 0:
		go func() {
			time.Sleep(ke.timeout)
			ke.cleanupIfHasError()
		}()
	}
}

func (ke *KitEndpointer) discoveryCallback(_ discovery.Instancer) {
	insts, e := ke.instancer.Instances(ke.selector)

	ke.mtx.Lock()
	defer ke.mtx.Unlock()

	if e != nil {
		ke.handleError(e)
		return
	}

	// prepare endpoints
	entries := map[string]endpointEntry{}
	for _, inst := range insts {
		if entry, ok := ke.entries[inst.ID]; ok {
			entries[inst.ID] = entry
		} else {
			// create endpoint
			ep, closer, e := ke.factory(inst)
			if e != nil {
				ke.handleError(e)
				return
			}
			entries[inst.ID] = endpointEntry{
				ep:     ep,
				closer: closer,
			}
		}
	}

	// cleanup endpoints
	ke.entries = entries
	ke.lastErr = nil
	ke.endpoints = make([]endpoint.Endpoint, len(insts))
	var i int
	for _, entry := range ke.entries {
		ke.endpoints[i] = entry.ep
		i++
	}
}


