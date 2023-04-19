package access

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"fmt"
	"strings"
)

/**************************
	Common ControlFunc
***************************/

// HasPermissions returns a ControlFunc that checks permissions of current auth.
// If the given auth doesn't contain all specified permission, the ControlFunc returns false and a security.AccessDeniedError
func HasPermissions(permissions...string) ControlFunc {
	return func(auth security.Authentication) (bool, error) {
		switch {
		case auth.State() > security.StateAnonymous && security.HasPermissions(auth, permissions...):
			return true, nil
		case auth.State() < security.StatePrincipalKnown:
			return false, security.NewInsufficientAuthError("not authenticated")
		case auth.State() < security.StateAuthenticated:
			return false, security.NewInsufficientAuthError("not fully authenticated")
		default:
			return false, security.NewAccessDeniedError("access denied")
		}
	}
}

/**************************
	Permission Expr
***************************/

const (
	opAnd   = "&&"
	opOr    = "||"
	opNot   = "!"
	opOpen  = "("
	opClose = ")"
)

// HasPermissionsWithExpr takes an expression and returns a ControlFunc that evaluate security.Permissions against
// the given expression.
//
// The expression is composed by 1 or more expression-unit combined using logical operands and brackets.
// supported expresion-unit are:
// 	- !<permission>
// 	- <permission> && <permission>
// 	- <permission> || <permission>
// where <permission> stands for "security.Permissions.Has(<permission>)" which yields bool result
// e.g. "P1 && P2 && !(P3 || P4)", means security.Permissions contains both P1 and P2 but not contains neither P3 nor P4
func HasPermissionsWithExpr(expr string) ControlFunc {

	matcher, e := parsePermissionExpr(expr)
	if matcher == nil {
		expr = strings.ReplaceAll(expr, " ", "")
		panic(fmt.Errorf(`Invalid permission expression "%s": %v`, expr, e))
	}

	return func(auth security.Authentication) (bool, error) {
		if auth.State() > security.StateAnonymous {
			//user with API admin permission is allowed to short cut the permission check
			if security.HasPermissions(auth, security.SpecialPermissionAPIAdmin) {
				return true, nil
			}

			if match, e := matcher.Matches(auth.Permissions()); match && e == nil {
				return true, nil
			}
		}

		switch {
		case auth.State() < security.StatePrincipalKnown:
			return false, security.NewInsufficientAuthError("not authenticated")
		case auth.State() < security.StateAuthenticated:
			return false, security.NewInsufficientAuthError("not fully authenticated")
		default:
			return false, security.NewAccessDeniedError("access denied")
		}
	}
}

/**************************
	Expr Parsing Helpers
***************************/
var opTokens = utils.NewStringSet(opAnd, opOr, opNot, opOpen, opClose)

type operand struct {
	op    string
	order int
}

func operandFromString(str string) operand {
	switch str {
	case opOr:
		return operand{op:opOr, order: 1}
	case opAnd:
		return operand{op:opAnd, order: 2}
	case opNot:
		return operand{op:opNot, order: 3}
	default:
		return operand{op:"", order: 0}
	}
}

func (o operand) Precedence(p int) operand {
	return operand{
		op: o.op,
		order: o.order + p,
	}
}

func (o operand) String() string {
	return o.op
}

// parsePermissionExpr takes an expression and parse it into a matcher.ChainableMatcher
// it parse the expr with helps of two FILO stacks ("ops" and "args"):
// - "ops" holds all operands to be processed ("!", "&&", "||"), each has a precedence/order representing evaluation priority
// - "args" holds permissions matchers to be processed
// - "args" will eventually be reduced into a single matcher.ChainableMatcher that represent the overall expression
// see processOperand for more details
func parsePermissionExpr(expr string) (ret matcher.Matcher, err error) {
	expr = strings.ReplaceAll(expr, " ", "")
	var ops []operand
	var args []matcher.Matcher

	lastToken := ""
	var precedence, idx int
	for remaining := expr; remaining != ""; {
		t, r := nextToken(remaining)
		if opTokens.Has(t) {
			switch t {
			case opOpen:
				precedence = precedence + 10
			case opClose:
				if precedence == 0 {
					return nil, fmt.Errorf(`found ")" without matching "(" at idx %d`, idx)
				}
				precedence = precedence - 10
			case opOr, opAnd:
				if opTokens.Has(lastToken) {
					// we have && or || follows another operand, this is invalid. e.g. "A && || B"
					return nil, fmt.Errorf(`found "&&" or "||" following another operand at idx %d`, idx)
				}
				fallthrough
			default:
				op := operandFromString(t).Precedence(precedence)
				ops, args = processOperand(ops, args, op)
				if op.op == "" || ops == nil {
					return nil, fmt.Errorf(`unexpected error at idx %d`, idx)
				}
			}
		} else {
			if strings.ContainsAny(t, "&|!()") {
				return nil, fmt.Errorf(`invalid permission value idx %d`, idx)
			}
			args = append(args, NewPermissionMatcher(t))
		}
		idx = idx + len(t)
		remaining = r
	}

	if precedence != 0 {
		// we don't have matching number of "(" and ")"
		return nil, fmt.Errorf(`unexpected EOF, found "(" without matching ")"`)
	}

	ops, args = processOperand(ops, args, operand{})
	if len(ops) != 1 || len(args) != 1 {
		return nil, fmt.Errorf(`unexpected EOF, unknown error`)
	}
	ret = args[0]
	return
}

// processOperand takes the new operand and existing operand stack and args stack, and returns the processed stacks
// using rule:
// 1. ops and args are stack, and FILO
// 2. all elements in ops stack should also be in ASC order
// 3.1. if ops stack is empty, OR the newOp have higher order than any ops in the stack
//      the newOp is pushed into ops stack
// 3.2  otherwise, existing ops stack should be reduced by poping top ops with same operand/order
//      and combining corresponding args at stack top into single value
// 3.3 repeat 3.1 & 3.2 untils condition 2 is satisfied
//
// e.g.
// 	A || !(B || C && ! D) || !E || ! !F
//	  1  3   11   12 13   1  3  1  3 3
//                        ^
// 	A || !(B || C && ?) || !E || ! !F	[? = !D]
//	  1  3   11   12    1  3  1  3 3
//                      ^
// 	A || !(B || ?) || !E || ! !F	[? = C && !D]
//	  1  3   11    1  3  1  3 3
//                 ^
// 	A || ! ? || !E || ! !F	[? = B || C && !D]
//	  1  3   1  3  1  3 3
//           ^
//  Done, move on to next operand
// 	A || ? || !E || ! !F	[? = !(B || C && !D)]
//	  1    1  3  1  3 3
//            ^
// ...
func processOperand(ops []operand, args []matcher.Matcher, newOp operand) ([]operand, []matcher.Matcher) {
	// special case: empty ops
	if len(ops) == 0 {
		return append(ops, newOp), args
	}

	// repeat 3.1, 3.2 until all elements in ops stack have lower order than newOp
	for {
		// terminal condition: if top of the ops stack is not higher order than the target, we are done
		if len(ops) == 0 || ops[len(ops) - 1].order <= newOp.order {
			break
		}

		// start with last arg
		last := ops[len(ops) - 1]
		arg := len(args) - 1
		var opCount int
		for i := len(ops) - 1; i >= 0 && ops[i].order == last.order; i-- {
			opCount ++
			op := ops[i].op
			if op != opNot {
				// two arguments required, will reduce one more arg
				arg --
			}

			if op != last.op || arg < 0 {
				// 1. for each percedence, we only have one operand: either ||, $$ or !.
				// 	  so if we find something have different operand but same precedence, there must be samething wrong
				// 2. not enough args
				return nil, nil
			}
		}

		// combine top of args stack into single matcher using the operand at the top operand stack
		ops, args = combine(ops, args, len(ops) - opCount, arg)
	}

	// push newOp and return
	ops = append(ops, newOp)
	return ops, args
}

// combine N args at top of "args" stack into single arg, where N is specified via "aIdx".
// also remove M operands
// combine top of the stacks ("operands" starting at opIdx and "args" starting at aIdx)
func combine(ops []operand, args []matcher.Matcher, opIdx, aIdx int) ([]operand, []matcher.Matcher) {
	op := ops[opIdx]
	combined := args[aIdx]
	switch op.op {
	case opNot:
		if (len(ops) - opIdx) % 2 == 1 {
			combined = matcher.Not(combined)
		}
	case opOr:
		combined = matcher.Or(args[aIdx], args[aIdx + 1:]...)
	case opAnd:
		combined = matcher.And(args[aIdx], args[aIdx + 1:]...)
	}
	ops = ops[:opIdx]
	args = append(args[:aIdx], combined)
	return ops, args
}

// nextToken assume there is no space in string
func nextToken(str string) (token string, remaining string) {
	for i, _ := range str {
		t := str[:i+1]
		match, op := isEndWithOp(t)
		if match {
			idx := i + 1 - len(op)
			if idx == 0 {
				// the string begin with op
				return op, str[i+1:]
			}
			// return string before the op
			return t[:idx], str[idx:]
		}
	}
	return str, ""
}

func isEndWithOp(str string) (match bool, op string) {
	switch l := len(str); {
	case l < 1:
		return false, ""
	case l < 2:
		switch t := str[l-1:]; t {
		case opNot, opOpen, opClose:
			return true, t
		}
	default:
		switch t := str[l-1:]; t {
		case opNot, opOpen, opClose:
			return true, t
		}
		switch t := str[l-2:]; t {
		case opAnd, opOr:
			return true, t
		}
	}
	return false, ""
}

/**************************
	Permission Matcher
***************************/
// permissionMatcher implements matcher.ChainableMatcher and accept map[string]interface{}
type permissionMatcher struct {
	permission string
}

func NewPermissionMatcher(permission string) *permissionMatcher {
	return &permissionMatcher{
		permission: permission,
	}
}

func (m *permissionMatcher) Matches(i interface{}) (bool, error) {
	if perms, ok := i.(security.Permissions); ok {
		return perms.Has(m.permission), nil
	}
	return false, nil
}

func (m *permissionMatcher) MatchesWithContext(_ context.Context, i interface{}) (bool, error) {
	return m.Matches(i)
}

func (m *permissionMatcher) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *permissionMatcher) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

func (m *permissionMatcher) String() string {
	return m.permission
}



