package tools

import (
	"fmt"

	"github.com/bblfsh/go-client/v4/tools"
	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/role"
)

// The cyclomatic complexity is a quantitative measure of the number of linearly
// independent paths through a program's source code.
// It was developed by Thomas J. McCabe, Sr. in 1976. For a formal description see:
// https://en.wikipedia.org/wiki/Cyclomatic_complexity
// And the original paper: http://www.literateprogramming.com/mccabe.pdf

// This implementation uses PMD implementation as reference and uses the method of
// counting one + one of the following UAST Roles if present on any children:
// * Statement, If | Case | For | While | DoWhile | Continue
// * Try, Catch
// * Operator, Boolean
// * Goto
// Important: since some languages allow for code defined
// outside function definitions, this won't check that the Node has the roles Function, Declaration
// so the user should check that if the intended use is calculating the complexity of a function/method.
// If the children contain more than one function definitions, the value will not be averaged between
// the total number of function declarations but given as a total.
//
// Some practical implementations counting tokens in the code. They sometimes differ; for example
// some of them count the switch "default" as an incrementor, some consider all return values minus the
// last, some of them consider "else" (which is wrong IMHO, but not for elifs, remember than the IfElse
// token in the UAST is really an Else not an "else if", elseifs would have a children If token), some
// consider throw and finally while others only the catch, etc.
//
// Examples:
// PMD reference implementation: http://pmd.sourceforge.net/pmd-4.3.0/xref/net/sourceforge/pmd/rules/CyclomaticComplexity.html
// GMetrics: http://gmetrics.sourceforge.net/gmetrics-CyclomaticComplexityMetric.html
// Go: https://github.com/fzipp/gocyclo/blob/master/gocyclo.go#L214
// SonarQube (include rules for many languages): https://docs.sonarqube.org/display/SONAR/Metrics+-+Complexity
//
// IMPORTANT DISCLAIMER: McCabe definition specifies clearly that boolean operations should increment the
// count in 1 for every boolean element when the language if the language evaluates conditions in
// short-circuit.  Unfortunately in the current version of the UAST we don't specify these language invariants
// and also we still haven't defined how we are going to represent the boolean expressions (which also would
// need a tree transformation process in the pipeline that we lack) so lacking a better way of detecting for
// all  languages what symbols or literals are part of a boolean expression we count the boolean operators
// themselves which should work for short-circuit infix languages but not prefix or infix languages that can
// evaluate more than two items with a single operator.  (FIXME when both things are solved in the UAST
// definition and the SDK).

// CyclomaticComplexity is a sub-command that compuets cyclomatic complexity.
type CyclomaticComplexity struct{}

func (cc CyclomaticComplexity) Exec(n nodes.Node) error {
	result := cyclomaticComplexity(n)
	fmt.Println("Cyclomatic Complexity = ", result)
	return nil
}

// cyclomaticComplexity returns the cyclomatic complexity for the node.
func cyclomaticComplexity(n nodes.Node) int {
	complexity := 1

	iter := tools.NewIterator(n, tools.PreOrder)

	for n := range tools.Iterate(iter) {
		roles := make(map[role.Role]bool)
		for _, r := range uast.RolesOf(n) {
			roles[r] = true
		}
		if addsComplexity(roles) {
			complexity++
		}
	}
	return complexity
}

func addsComplexity(roles map[role.Role]bool) bool {
	return roles[role.Statement] && (roles[role.If] || roles[role.Case] || roles[role.For] || roles[role.While] || roles[role.DoWhile] || roles[role.Continue]) ||
		roles[role.Try] && roles[role.Catch] ||
		roles[role.Operator] && roles[role.Boolean] ||
		roles[role.Goto]
}
