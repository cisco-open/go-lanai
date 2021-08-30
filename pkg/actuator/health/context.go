package health

import (
	"context"
	"strings"
)

const (
	StatusUnkown Status = iota
	StatusUp
	StatusOutOfService
	StatusDown
)

type Status int

// fmt.Stringer
func (s Status) String() string {
	switch s {
	case StatusUp:
		return "UP"
	case StatusDown:
		return "DOWN"
	case StatusOutOfService:
		return "OUT_OF_SERVICE"
	default:
		return "UNKNOWN"
	}
}

// encoding.TextMarshaler
func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// encoding.TextUnmarshaler
func (s *Status) UnmarshalText(data []byte) error {
	value := strings.ToUpper(string(data),)
	switch value {
	case "UP":
		*s = StatusUp
	case "DOWN":
		*s = StatusDown
	case "OUT_OF_SERVICE":
		*s = StatusOutOfService
	default:
		*s = StatusUnkown
	}
	return nil
}

const (
	// Never show the item in the response.
	ShowModeNever ShowMode = iota
	// Show the item in the response when accessed by an authorized user.
	ShowModeAuthorized
	// Always show the item in the response.
	ShowModeAlways
)

// ShowMode is options for showing items in responses from the HealthEndpoint web extensions.
type ShowMode int

// fmt.Stringer
func (m ShowMode) String() string {
	switch m {
	case ShowModeAuthorized:
		return "authorized"
	case ShowModeAlways:
		return "always"
	default:
		return "never"
	}
}

// encoding.TextMarshaler
func (s ShowMode) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// encoding.TextUnmarshaler
func (m *ShowMode) UnmarshalText(data []byte) error {
	value := strings.ToLower(string(data))
	switch value {
	case "authorized", "when_authorized", "whenAuthorized", "when-authorized":
		*m = ShowModeAuthorized
	case "always":
		*m = ShowModeAlways
	default:
		*m = ShowModeNever
	}
	return nil
}

type StatusAggregator interface {
	Aggregate(context.Context, ...Status) Status
}

type StatusCodeMapper interface {
	StatusCode(context.Context, Status) int
}

type Health interface {
	Status() Status
	Description() string
}

type Options struct {
	ShowDetails    bool
	ShowComponents bool
}

type Indicator interface {
	Name() string
	Health(context.Context, Options) Health
}