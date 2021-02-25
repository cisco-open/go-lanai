package health

import (
	"context"
	"net/http"
)

/*******************************
	StaticStatusCodeMapper
********************************/
var DefaultStaticStatusCodeMapper = StaticStatusCodeMapper{
	StatusUp:           http.StatusOK,
	StatusDown:         http.StatusServiceUnavailable,
	StatusOutOfService: http.StatusServiceUnavailable,
	StatusUnkown:       http.StatusInternalServerError,
}

type StaticStatusCodeMapper map[Status]int

func (m StaticStatusCodeMapper) StatusCode(_ context.Context, status Status) int {
	if sc, ok := m[status]; ok {
		return sc
	}
	return http.StatusServiceUnavailable
}

/*******************************
	SimpleHealth
********************************/
// SimpleHealth implements Health
type SimpleHealth struct {
	Stat Status `json:"status"`
	Desc string `json:"description,omitempty"`
}

func (h SimpleHealth) Status() Status {
	return h.Stat
}

func (h SimpleHealth) Description() string {
	return h.Desc
}

/*******************************
	Composite
********************************/
// CompositeHealth implement Health
type CompositeHealth struct {
	SimpleHealth
	Components map[string]Health `json:"components,omitempty"`
}

func NewCompositeHealth(status Status, description string, components map[string]Health) *CompositeHealth {
	return &CompositeHealth{
		SimpleHealth: SimpleHealth{
			Stat: status,
			Desc: description,
		},
		Components: components,
	}
}

/*******************************
	DetailedHealth
********************************/
// DetailedHealth implement Health
type DetailedHealth struct {
	SimpleHealth
	Details map[string]interface{} `json:"details,omitempty"`
}

func NewDetailedHealth(status Status, description string, details map[string]interface{}) *DetailedHealth {
	return &DetailedHealth{
		SimpleHealth: SimpleHealth{
			Stat: status,
			Desc: description,
		},
		Details: details,
	}
}
