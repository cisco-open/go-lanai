package discoverymock

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"strconv"
	"strings"
)

func BeHealthy() InstanceMockOptions {
	return func(inst *discovery.Instance) {
		inst.Health = discovery.HealthPassing
	}
}

func BeCritical() InstanceMockOptions {
	return func(inst *discovery.Instance) {
		inst.Health = discovery.HealthCritical
	}
}

func WithExtraTag(tags ...string) InstanceMockOptions {
	return func(inst *discovery.Instance) {
		inst.Tags = append(inst.Tags, tags...)
	}
}

func WithMeta(k, v string) InstanceMockOptions {
	return func(inst *discovery.Instance) {
		inst.Meta[k] = v
	}
}

func AnyInstance() InstanceMockMatcher {
	return func(inst *discovery.Instance) bool {
		return true
	}
}

func NthInstance(n int) InstanceMockMatcher {
	return func(inst *discovery.Instance) bool {
		i := extractIndexIfPossible(inst)
		return i == n
	}
}

func InstanceAfterN(n int) InstanceMockMatcher {
	return func(inst *discovery.Instance) bool {
		i := extractIndexIfPossible(inst)
		return i > n
	}
}

func extractIndexIfPossible(inst *discovery.Instance) int {
	split := strings.SplitN(inst.ID, "-", 2)
	i, e := strconv.Atoi(split[0])
	if e != nil {
		return -1
	}
	return i
}