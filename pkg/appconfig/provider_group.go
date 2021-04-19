package appconfig

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"strings"
)

// StaticProviderGroup implements ProviderGroup, and holds fixed provider list
type StaticProviderGroup struct {
	Precedence      int
	StaticProviders []Provider
}

func NewStaticProviderGroup(order int, providers... Provider) *StaticProviderGroup {
	return &StaticProviderGroup{
		Precedence:      order,
		StaticProviders: providers,
	}
}

func (g StaticProviderGroup) Order() int {
	return g.Precedence
}

func (g StaticProviderGroup) Providers(_ context.Context, _ bootstrap.ApplicationConfig) []Provider {
	return g.StaticProviders
}

func (g *StaticProviderGroup) Reset() {
	for _, p := range g.StaticProviders {
		p.Reset()
	}
}

// DynamicProviderGroup implements ProviderGroup, and holds a sorted list of keys and their corresponding Provider.
// This type is typically used as embedded struct
type DynamicProviderGroup struct {
	Precedence        int
	ProviderKeys      []string // Provider should be sorted all time based on their provider's ordering
	ProviderLookup    map[string]Provider
	ResolvedProviders []Provider
	ProcessFunc       func(context.Context, []Provider) []Provider // ProcessFunc is invoked before setting ResolvedProviders. Last chance to change
}

func NewDynamicProviderGroup(order int) *DynamicProviderGroup {
	return &DynamicProviderGroup{
		Precedence:     order,
		ProviderKeys:   []string{},
		ProviderLookup: map[string]Provider{},
	}
}

func (g *DynamicProviderGroup) Order() int {
	return g.Precedence
}

func (g *DynamicProviderGroup) Providers(ctx context.Context, _ bootstrap.ApplicationConfig) (providers []Provider) {
	if g.ResolvedProviders != nil {
		return g.ResolvedProviders
	}

	// we assume ProviderKeys are sorted already
	// Note, we re-assign order of each providers starting with group's order and move backwards
	for i, order := len(g.ProviderKeys) - 1, g.Precedence; i >= 0; i-- {
		p, ok := g.ProviderLookup[g.ProviderKeys[i]]
		if !ok {
			continue
		}
		providers = append(providers, p)

		// re-assign order
		if ro, ok := p.(ProviderReorderer); ok {
			ro.Reorder(order)
		}
		order--
	}

	// process and return
	if g.ProcessFunc != nil {
		providers = g.ProcessFunc(ctx, providers)
	}
	g.ResolvedProviders = providers
	return
}

func (g *DynamicProviderGroup) Reset() {
	for _, p := range g.ProviderLookup {
		p.Reset()
	}
	g.ResolvedProviders = nil
}

// ProfileBasedProviderGroup extends DynamicProviderGroup and implements ProviderGroup
// it provide base methods to determine Providers based on PropertyKeyActiveProfiles
type ProfileBasedProviderGroup struct {
	DynamicProviderGroup
	KeyFunc    func(profile string) (key string)
	CreateFunc func(name string, order int, conf bootstrap.ApplicationConfig) Provider
}

func NewProfileBasedProviderGroup(order int) *ProfileBasedProviderGroup {
	return &ProfileBasedProviderGroup{
		DynamicProviderGroup: *NewDynamicProviderGroup(order),
	}
}

func (g *ProfileBasedProviderGroup) Providers(ctx context.Context, conf bootstrap.ApplicationConfig) (providers []Provider) {

	profiles := g.resolveProfiles(conf)

	// resolve names, create new providers if necessary
	g.ProviderKeys = []string{}
	names := map[string]struct{}{}
	lenBefore := len(g.ProviderLookup)
	for _, pf := range profiles {
		name := g.KeyFunc(pf)
		names[name] = struct{}{}
		g.ProviderKeys = append(g.ProviderKeys, name)

		if p, ok := g.ProviderLookup[name]; !ok || p == nil {
			p = g.CreateFunc(name, g.Precedence, conf)
			if p != nil {
				g.ProviderLookup[name] = p
			}
		}
	}

	// cleanup ProviderLookup to prevent mem leak
	if lenBefore != len(g.ProviderLookup) {
		for k := range g.ProviderLookup {
			if _, ok := names[k]; !ok {
				delete(g.ProviderLookup, k)
			}
		}
		// reset resolved providers too
		g.ResolvedProviders = nil
	}

	return g.DynamicProviderGroup.Providers(ctx, conf)
}

func (g *ProfileBasedProviderGroup) resolveProfiles(conf bootstrap.ApplicationConfig) (profiles []string) {
	// active profiles
	active, _ := conf.Value(PropertyKeyActiveProfiles).([]string)
	for _, p := range active {
		p = strings.TrimSpace(p)
		if p != "" {
			profiles = append(profiles, p)
		}
	}

	// additional profiles
	additional, _ := conf.Value(PropertyKeyAdditionalProfiles).([]string)
	for _, p := range additional {
		p = strings.TrimSpace(p)
		if p != "" {
			profiles = append(profiles, p)
		}
	}
	// default profiles
	profiles = append(profiles, "") // add default profile
	return
}
