package discovery

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"strings"
)

type Customizer interface {
	Customize(ctx context.Context, reg *api.AgentServiceRegistration)
}

type Customizers struct {
	Customizers utils.Set
	applied bool
}

func NewCustomizers(ctx *bootstrap.ApplicationContext) *Customizers {
	return &Customizers{
		Customizers: utils.NewSet(NewDefaultCustomizer(ctx), buildInfoDiscoveryCustomizer{}),
	}
}

func (r *Customizers) Add(c Customizer) {
	if r.applied {
		panic(errors.New("cannot add consul registration customizer because other customization has already been applied"))
	}
	r.Customizers.Add(c)
}

func (r *Customizers) Apply(ctx context.Context, registration *api.AgentServiceRegistration) {
	if r.applied {
		return
	}
	defer func() {
		r.applied = true
	}()

	for c, _ := range r.Customizers {
		c.(Customizer).Customize(ctx, registration)
	}
}

type buildInfoDiscoveryCustomizer struct {}

func (b buildInfoDiscoveryCustomizer) Customize(ctx context.Context, reg *api.AgentServiceRegistration) {
	attrs := map[string]string {
		TAG_VERSION: bootstrap.BuildVersion,
		TAG_BUILD_DATE_TIME: bootstrap.BuildTime,
	}

	components := strings.Split(bootstrap.BuildVersion, "-")
	if len(components) == 2 {
		attrs[TAG_BUILD_NUMBER] = components[1]
	}

	if reg.Meta == nil {
		reg.Meta = map[string]string{}
	}

	for k, v := range attrs {
		reg.Meta[k] = v
		reg.Tags = append(reg.Tags, fmt.Sprintf("%s=%s", k, v))
	}
}