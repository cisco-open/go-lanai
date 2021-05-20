package discovery

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"fmt"
	"strings"
	"time"
)

const (
	InstanceMetaKeyVersion = "version"
	InstanceMetaKeyContextPath = "context"
	InstanceMetaKeySMCR = "SMCR"
	InstanceMetaKeySecure = "secure"
	//InstanceMetaKey = ""
)

const (
	HealthAny HealthStatus = iota
	HealthPassing
	HealthWarning
	HealthCritical
	HealthMaintenance
)

var (
	ErrInstancerStopped = fmt.Errorf("instancer is already stopped")
)

type Client interface {
	Instancer(serviceName string) (Instancer, error)
}

// HealthStatus maintenance > critical > warning > passing
type HealthStatus int

type Service struct {
	Name       string
	Insts      []*Instance
	Time       time.Time
	Err        error
	FirstErrAt time.Time
}

func (s *Service) Instances(selector InstanceMatcher) (ret []*Instance) {
	for _, inst := range s.Insts {
		if selector != nil {
			if matched, e := selector.Matches(inst); e != nil || !matched {
				continue
			}
		}
		ret = append(ret, inst)
	}
	return
}

type Instance struct {
	ID       string
	Service  string
	Address  string
	Port     int
	Tags     []string
	Meta     map[string]string
	Health   HealthStatus
	RawEntry interface{}
}

type Callback func(Instancer)

// InstanceMatcher is a matcher.Matcher that takes Instance or *Instance
type InstanceMatcher matcher.ChainableMatcher

type Instancer interface {
	ServiceName() string
	Service() *Service
	Instances(InstanceMatcher) ([]*Instance, error)
	Start(ctx context.Context)
	Stop()
	RegisterCallback(id interface{}, cb Callback)
	DeregisterCallback(id interface{})
}

// ServiceCache is not goroutine-safe unless the detail implementation says so
type ServiceCache interface {
	// Get returns service with given service name. return nil if not exist
	Get(name string) *Service
	// Set stores given service with name, returns non-nil if the service is already exists
	Set(name string, svc *Service) *Service
	// SetWithTTL stores given service with name and TTL, returns non-nil if the service is already exists
	// if ttl is zero or negative value, it's equivalent to Set
	SetWithTTL(name string, svc *Service, ttl time.Duration) *Service
	Has(name string) bool
	Entries() map[string]*Service
}

/*************************
	Common Impl
 *************************/

var (
	healthyInstanceMatcher = &instanceMatcher{
		desc:      "is healthy",
		matchFunc: func(_ context.Context, instance *Instance) (bool, error) {
			return instance.Health == HealthPassing, nil
		},
	}
)

// InstanceIsHealthy returns an InstanceMatcher that matches healthy instances
func InstanceIsHealthy() InstanceMatcher {
	return healthyInstanceMatcher
}

func InstanceWithVersion(verPattern string) InstanceMatcher {
	return &instanceMatcher{
		desc:      "is healthy",
		matchFunc: func(_ context.Context, instance *Instance) (bool, error) {
			if instance.Meta == nil {
				return false, nil
			}
			ver, ok := instance.Meta[InstanceMetaKeyVersion]
			return ok && ver == verPattern, nil
		},
	}
}

func InstanceWithTagKV(key, value string, caseInsensitive bool) InstanceMatcher {
	if caseInsensitive {
		key = strings.ToLower(key)
		value = strings.ToLower(value)
	}

	return &instanceMatcher{
		desc:      fmt.Sprintf("with tag %s=%s", key, value),
		matchFunc: func(_ context.Context, instance *Instance) (bool, error) {
			if instance.Tags == nil {
				return false, nil
			}
			for _, tag := range instance.Tags {
				if caseInsensitive {
					tag = strings.ToLower(tag)
				}
				kv := strings.SplitN(strings.TrimSpace(tag), "=", 2)
				if len(kv) == 2 && kv[0] == key && kv[1] == value {
					return true, nil
				}
			}
			return false, nil
		},
	}
}

// instanceMatcher implements InstanceMatcher and accept *Instance and Instance
type instanceMatcher struct {
	matchFunc func(context.Context, *Instance) (bool, error)
	desc      string
}

func (m *instanceMatcher) Matches(i interface{}) (bool, error) {
	return m.MatchesWithContext(context.TODO(), i)
}

func (m *instanceMatcher) MatchesWithContext(c context.Context, i interface{}) (ret bool, err error) {
	var inst *Instance
	switch v := i.(type) {
	case *Instance:
		inst = v
	case Instance:
		inst = &v
	default:
		return false, fmt.Errorf("expect *Instance but got %T", i)
	}
	return m.matchFunc(c, inst)
}

func (m *instanceMatcher) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *instanceMatcher) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

func (m *instanceMatcher) String() string {
	return m.desc
}