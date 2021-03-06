package access

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"fmt"
)

// ControlCondition extends web.RequestMatcher, and matcher.ChainableMatcher
// it is used together with web.RoutedMapping's "Condition" for a convienent config of securities
// only matcher.ChainableMatcher's .MatchesWithContext (context.Context, interface{}) (bool, error) is used
// Matches(interface{}) (bool, error) should return regular as if the context is empty
//
// In addition, implementation should also return AccessDeniedError when condition didn't match.
// web.Registrar will propagate this error along the handler chain until it's handled by errorhandling middleware
type ControlCondition matcher.ChainableMatcher

/**************************
	Common Impl.
***************************/
type matchableFunc func(context.Context, interface{}) (interface{}, error)

type controlCondition struct {
	description string
	controlFunc ControlFunc
}

func (m *controlCondition) Matches(i interface{}) (bool, error) {
	return m.MatchesWithContext(context.TODO(), i)
}

func (m *controlCondition) MatchesWithContext(c context.Context, _ interface{}) (bool, error) {
	auth := security.Get(c)
	return m.controlFunc(auth)
}

func (m *controlCondition) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *controlCondition) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

func (m controlCondition) String() string {
	switch {
	case len(m.description) != 0:
		return m.description
	default:
		return "access.ControlCondition"
	}
}

/**************************
	Constructors
***************************/
// RequirePermissions returns ControlCondition using HasPermissionsWithExpr
// e.g. RequirePermissions("P1 && P2 && !(P3 || P4)"), means security.Permissions contains both P1 and P2 but not contains neither P3 nor P4
// see HasPermissionsWithExpr for expression syntax
func RequirePermissions(expr string) ControlCondition {
	return &controlCondition{
		description:   fmt.Sprintf("user's permissions match [%s]", expr),
		controlFunc:   HasPermissionsWithExpr(expr),
	}
}

/**************************
	Helpers
***************************/

