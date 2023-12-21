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

package access

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	. "github.com/onsi/gomega"
	"testing"
)

/*************************
	Test Cases
 *************************/
func TestValidPermissionExprParsing(t *testing.T) {
	t.Run("SingleTest", ValidExprParseTest("A", [][]string{
		// should match
		[]string{"A"}, []string{"A", "B"},
	}, [][]string{
		// should not match
		[]string{"B"}, []string{},
	}))

	t.Run("SingleNotTest", ValidExprParseTest("!A", [][]string{
		// should match
		[]string{"B"}, []string{},
	}, [][]string{
		// should not match
		[]string{"A"}, []string{"A", "B"},
	}))

	t.Run("SingleAndTest", ValidExprParseTest("A && B", [][]string{
		// should match
		[]string{"A", "B"},
	}, [][]string{
		// should not match
		[]string{"B"}, []string{"A"}, []string{}, []string{"C", "A"}, []string{"C", "B"}, []string{"C"},
	}))

	t.Run("MultiOrTest", ValidExprParseTest("A || B || C", [][]string{
		// should match
		[]string{"A", "B"}, []string{"B"}, []string{"A"}, []string{"C", "A"}, []string{"C", "B"}, []string{"C"},
	}, [][]string{
		// should not match
		[]string{}, []string{"D"},
	}))

	t.Run("MultiAndTest", ValidExprParseTest("A && B && C", [][]string{
		// should match
		[]string{"A", "B", "C"},
	}, [][]string{
		// should not match
		[]string{"A", "B"}, []string{"B"}, []string{"A"}, []string{"C", "A"}, []string{"C", "B"},
		[]string{}, []string{"C"},
	}))

	t.Run("AndNotTest", ValidExprParseTest("A && !B", [][]string{
		// should match
		[]string{"A", "C"}, []string{"A"},
	}, [][]string{
		// should not match
		[]string{}, []string{"C"}, []string{"B"},
		[]string{"A", "B"}, []string{"C", "B"},
	}))

	t.Run("OrNotTest", ValidExprParseTest("!A || B", [][]string{
		// should match
		[]string{}, []string{"C"},
		[]string{"A", "B"}, []string{"B"},
	}, [][]string{
		// should not match
		[]string{"A"}, []string{"A", "C"},
	}))

	t.Run("AndOrNotTest", ValidExprParseTest("C || A && !B", [][]string{
		// should match
		[]string{"C"}, []string{"C", "B"}, []string{"A", "B", "C"},
		[]string{"A", "C"}, []string{"A"},
	}, [][]string{
		// should not match
		[]string{},  []string{"B"},
		[]string{"A", "B"},
	}))

	// More complex test cases
	// Case A1 & A2: "A || !(B || C && !D) || !E && !!F" == "A || !B && !C || !B && D || !E && F"
	// exprA1 and exprA2 are equevalent
	exprA1 := "A || !(B || C && !D) || !E && !!F"
	exprA2 := "A || !B && !C || !B && D || !E && F"
	possitiveA := [][]string{
		// should match
		[]string{"A"}, []string{"A", "B"}, []string{"A", "C"}, []string{"A", "D"}, []string{"A", "D", "E", "F"},
		[]string{"E", "F"}, []string{"D", "E", "F"}, []string{"D", "F"}, []string{"D", "E"},
		[]string{"C", "D"}, []string{"C", "D", "F"}, []string{"D", "C", "E", "F"}, []string{"D", "C", "E"},
		[]string{"F"}, []string{},
	}
	negativeA := [][]string{
		// should not match
		[]string{"B"}, []string{"B", "E", "F"}, []string{"B", "E"},
		[]string{"C"}, []string{"C", "E", "F"}, []string{"C", "E"},
		[]string{"B", "C", "D"}, []string{"B", "C", "D", "E", "F"}, []string{"B", "C", "D", "E"},
	}
	t.Run("ComplexExprATest", ValidExprParseTest(exprA1, possitiveA, negativeA))
	t.Run("ComplexExprAEquvlentTest", ValidExprParseTest(exprA2, possitiveA, negativeA))

	// Case B1 & B2: "A && (!!(B && C || D) && (!E || !F))"
	//	== "A && B && C && !E || A && B && C && !F || A && D && !E || A && D && !F"
	// exprB1 and exprB2 are equevalent
	exprB1 := "A && (!!(B && C || D) && (!E || !F))"
	exprB2 := "A && B && C && !E || A && B && C && !F || A && D && !E || A && D && !F"
	possitiveB := [][]string{
		// should match
		[]string{"A", "B", "C", "F"}, []string{"A", "B", "C"}, []string{"A", "B", "C", "D", "F"}, []string{"A", "B", "C", "D"},
		[]string{"A", "B", "C", "E"}, []string{"A", "B", "C", "E", "D"},
		[]string{"A", "C", "D", "F"}, []string{"A", "B", "D", "F"}, []string{"A", "C", "D"}, []string{"A", "B", "D"},
		[]string{"A", "C", "D", "E"}, []string{"A", "B", "D", "E"},
	}
	negativeB := [][]string{
		// should not match (too many cases, we can't list all)
		[]string{}, []string{"B"},
		[]string{"B", "C"}, []string{"B", "C", "D"}, []string{"B", "C", "D", "E"}, []string{"B", "C", "D", "E", "F"},
		[]string{"B", "D"}, []string{"B", "D", "E"}, []string{"B", "D", "E", "F"},
		[]string{"B", "E"}, []string{"B", "E", "F"}, []string{"B", "F"},
		[]string{"C"}, []string{"C", "D"}, []string{"C", "D", "E"}, []string{"C", "D", "E", "F"},
		[]string{"C", "E"}, []string{"C", "E", "F"}, []string{"C", "F"},
		[]string{"D"}, []string{"D", "E"}, []string{"D", "E", "F"}, []string{"D", "F"},

		[]string{"A", "D", "E", "F"}, []string{"A", "E"}, []string{"A", "F"}, []string{"A"},
		[]string{"A", "C", "D", "E", "F"}, []string{"A", "C", "E"}, []string{"A", "C", "F"}, []string{"A", "C"},
		[]string{"A", "B", "D", "E", "F"}, []string{"A", "B", "E"}, []string{"A", "B", "F"}, []string{"A", "B"},
		[]string{"A", "B", "C", "E", "F"}, []string{"A", "B", "C", "E", "F", "D"},
	}
	t.Run("ComplexExprATest", ValidExprParseTest(exprB1, possitiveB, negativeB))
	t.Run("ComplexExprAEquvlentTest", ValidExprParseTest(exprB2, possitiveB, negativeB))
}

func TestInvalidPermissionExprParsing(t *testing.T) {
	t.Run("MissingOpenTest", InvalidExprParseTest("A && (B||C) && D) || E"))
	t.Run("MissingCloseTest", InvalidExprParseTest("A && ((B||C) || D && E"))
	t.Run("MissingNotArgTest1", InvalidExprParseTest("A && B && !"))
	t.Run("MissingNotArgTest2", InvalidExprParseTest("A && ! && C"))
	t.Run("MissingOrArgTest1", InvalidExprParseTest("A && !B || "))
	t.Run("MissingOrArgTest2", InvalidExprParseTest("A || && C"))
	t.Run("MissingOrArgTest3", InvalidExprParseTest("|| B && C"))
	t.Run("MissingAndArgTest1", InvalidExprParseTest("A || !B && "))
	t.Run("MissingAndArgTest2", InvalidExprParseTest("A && && C"))
	t.Run("MissingAndArgTest3", InvalidExprParseTest("&& B || C"))
	t.Run("ShortOperandTest1", InvalidExprParseTest("A & B || C"))
	t.Run("ShortOperandTest2", InvalidExprParseTest("A && B | C"))
}

/*************************
	Sub Tests
 *************************/
func ValidExprParseTest(expr string, positive, negative [][]string) func(*testing.T) {
	return func(t *testing.T) {
		g := NewWithT(t)
		m, e := parsePermissionExpr(expr)
		g.Expect(e).To(Succeed(), "expr parse func should not return error with expr [%s]", expr)
		g.Expect(m).To(Not(BeNil()), "resulted matcher should not be nil with expr [%s]", expr)
		assertPositiveMatch(t, m, positive...)
		assertNegativeMatch(t, m, negative...)
	}
}

func InvalidExprParseTest(expr string) func(*testing.T) {
	return func(t *testing.T) {
		g := NewWithT(t)
		_, e := parsePermissionExpr(expr)
		g.Expect(e).To(Not(Succeed()), "expr parse func should return error with expr [%s]", expr)
	}
}

/*************************
	Helpers
 *************************/
func permissions(perms []string) security.Permissions {
	ret := security.Permissions{}
	for _, p := range perms {
		ret[p] = true
	}
	return ret
}

func assertPositiveMatch(t *testing.T, actual matcher.Matcher, conditions ...[]string) {
	g := NewWithT(t)
	for _, c := range conditions {
		p := permissions(c)
		r, e := actual.Matches(p)
		g.Expect(e).To(Succeed(), "matcher [%v] shouldn't return error on permissions %v", actual, p)
		g.Expect(r).To(BeTrue(), "matcher [%v] should match permissions %v", actual, p)
	}
}

func assertNegativeMatch(t *testing.T, actual matcher.Matcher, conditions ...[]string) {
	g := NewWithT(t)
	for _, c := range conditions {
		p := permissions(c)
		r, e := actual.Matches(p)
		g.Expect(e).To(Succeed(), "matcher [%v] shouldn't return error on permissions %v", actual, p)
		g.Expect(r).To(BeFalse(), "matcher [%v] should not match permissions %v", actual, p)
	}
}