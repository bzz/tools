package tools

import (
	"fmt"

	"github.com/bblfsh/go-client/v4/tools"
	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/role"
)

// NPath is a sub-command that computes NPath complexity.
type NPath struct{}

// NPathData represents a complexity for a single function.
type NPathData struct {
	Name       string
	Complexity int
}

func (np NPath) Exec(n nodes.Node) error {
	result := NPathComplexity(n)
	fmt.Println(result)
	return nil
}

func (nd *NPathData) String() string {
	return fmt.Sprintf("FuncName:%s, Complexity:%d\n", nd.Name, nd.Complexity)
}

// NPathComplexity computes the NPath of functions in a nodes.Node.
// PMD is considered the reference implementation to assert correctness.
// See: https://pmd.github.io/pmd-5.7.0/pmd-java/xref/net/sourceforge/pmd/lang/java/rule/codesize/NPathComplexityRule.html
func NPathComplexity(n nodes.Node) []*NPathData {
	var result []*NPathData
	var funcs []nodes.Node
	var names []string

	if containsRoles(n, []role.Role{role.Function, role.Body}, nil) {
		funcs = append(funcs, n)
		names = append(names, "NoName")
	} else {
		funcDecs := deepChildrenOfRoles(n, []role.Role{role.Function, role.Declaration}, []role.Role{role.Argument})
		for _, funcDec := range funcDecs {
			if containsRoles(funcDec, []role.Role{role.Function, role.Name}, nil) {
				names = append(names, uast.TokenOf(funcDec))
			}
			childNames := childrenOfRoles(funcDec, []role.Role{role.Function, role.Name}, nil)
			if len(childNames) > 0 {
				names = append(names, uast.TokenOf(childNames[0]))
			}
			childFuncs := childrenOfRoles(funcDec, []role.Role{role.Function, role.Body}, nil)
			if len(childFuncs) > 0 {
				funcs = append(funcs, childFuncs[0])
			}
		}
	}
	for i, function := range funcs {
		npath := visitFunctionBody(function)
		result = append(result, &NPathData{Name: names[i], Complexity: npath})
	}

	return result
}

func visitorSelector(n nodes.Node) int {
	if containsRoles(n, []role.Role{role.Statement, role.If}, []role.Role{role.Then, role.Else}) {
		return visitIf(n)
	}
	if containsRoles(n, []role.Role{role.Statement, role.While}, nil) {
		return visitWhile(n)
	}
	if containsRoles(n, []role.Role{role.Statement, role.Switch}, nil) {
		return visitSwitch(n)
	}
	if containsRoles(n, []role.Role{role.Statement, role.DoWhile}, nil) {
		return visitDoWhile(n)
	}
	if containsRoles(n, []role.Role{role.Statement, role.For}, nil) {
		return visitFor(n)
	}
	if containsRoles(n, []role.Role{role.Statement, role.Return}, nil) {
		return visitReturn(n)
	}
	if containsRoles(n, []role.Role{role.Statement, role.Try}, nil) {
		return visitTry(n)
	}
	return visitNotCompNode(n)
}

func complexityMultOf(n nodes.Node) int {
	npath := 1
	it, _ := tools.Filter(n, "/*")
	for it.Next() {
		child := it.Node().(nodes.Node)
		npath *= visitorSelector(child)
	}
	return npath
}

func visitFunctionBody(n nodes.Node) int {
	return complexityMultOf(n)
}

func visitNotCompNode(n nodes.Node) int {
	return complexityMultOf(n)
}

func visitIf(n nodes.Node) int {
	// (npath of if + npath of else (or 1) + bool_comp of if) * npath of next
	npath := 0
	ifThen := childrenOfRoles(n, []role.Role{role.If, role.Then}, nil)
	ifCondition := childrenOfRoles(n, []role.Role{role.If, role.Condition}, nil)
	ifElse := childrenOfRoles(n, []role.Role{role.If, role.Else}, nil)

	if len(ifElse) > 0 {
		npath += complexityMultOf(ifElse[0])
	} else {
		npath++
	}
	npath *= complexityMultOf(ifThen[0])
	npath += expressionComp(ifCondition[0])

	return npath
}

func visitWhile(n nodes.Node) int {
	// (npath of while + bool_comp of while + npath of else (or 1)) * npath of next
	npath := 0
	whileCondition := childrenOfRoles(n, []role.Role{role.While, role.Condition}, nil)
	whileBody := childrenOfRoles(n, []role.Role{role.While, role.Body}, nil)
	whileElse := childrenOfRoles(n, []role.Role{role.While, role.Else}, nil)
	// Some languages like python can have an else in a while loop
	if len(whileElse) > 0 {
		npath += complexityMultOf(whileElse[0])
	} else {
		npath++
	}

	npath *= complexityMultOf(whileBody[0])
	npath += expressionComp(whileCondition[0])

	return npath
}

func visitDoWhile(n nodes.Node) int {
	// (npath of do + bool_comp of do + 1) * npath of next
	npath := 1
	doWhileCondition := childrenOfRoles(n, []role.Role{role.DoWhile, role.Condition}, nil)
	doWhileBody := childrenOfRoles(n, []role.Role{role.DoWhile, role.Body}, nil)

	npath *= complexityMultOf(doWhileBody[0])
	npath += expressionComp(doWhileCondition[0])

	return npath
}

func visitFor(n nodes.Node) int {
	// (npath of for + bool_comp of for + 1) * npath of next
	npath := 1
	forBody := childrenOfRoles(n, []role.Role{role.For, role.Body}, nil)
	if len(forBody) > 0 {
		npath *= complexityMultOf(forBody[0])
	}
	npath++
	return npath
}

func visitReturn(n nodes.Node) int {
	if aux := expressionComp(n); aux != 1 {
		return aux - 1
	}
	return 1
}

func visitSwitch(n nodes.Node) int {
	caseDefault := childrenOfRoles(n, []role.Role{role.Switch, role.Default}, nil)
	switchCases := childrenOfRoles(n, []role.Role{role.Statement, role.Switch, role.Case}, []role.Role{role.Body})
	npath := 0

	if len(caseDefault) > 0 {
		npath += complexityMultOf(caseDefault[0])
	} else {
		npath++
	}
	for _, switchCase := range switchCases {
		npath += complexityMultOf(switchCase)
	}
	return npath
}

func visitTry(n nodes.Node) int {
	/*
		In pmd they decided the complexity of a try is the summatory of the complexity
		of the try body, catch body and finally body.I don't think this is the most acurate way
		of doing this.
	*/

	tryBody := childrenOfRoles(n, []role.Role{role.Try, role.Body}, nil)
	tryCatch := childrenOfRoles(n, []role.Role{role.Try, role.Catch}, nil)
	tryFinaly := childrenOfRoles(n, []role.Role{role.Try, role.Finally}, nil)

	catchComp := 0
	if len(tryCatch) > 0 {
		for _, catch := range tryCatch {
			catchComp += complexityMultOf(catch)
		}
	}
	finallyComp := 0
	if len(tryFinaly) > 0 {
		finallyComp = complexityMultOf(tryFinaly[0])
	}
	npath := complexityMultOf(tryBody[0]) + catchComp + finallyComp

	return npath
}

func visitConditionalExpr(n nodes.Node) {
	// TODO ternary operators are not defined on the UAST yet
}

func expressionComp(n nodes.Node) int {
	orCount := deepCountChildrenOfRoles(n, []role.Role{role.Operator, role.Boolean, role.And}, nil)
	andCount := deepCountChildrenOfRoles(n, []role.Role{role.Operator, role.Boolean, role.Or}, nil)

	return orCount + andCount + 1
}

func containsRoles(n nodes.Node, andRoles []role.Role, notRoles []role.Role) bool {
	roleMap := make(map[role.Role]bool)
	for _, r := range uast.RolesOf(n) {
		roleMap[r] = true
	}
	for _, r := range andRoles {
		if !roleMap[r] {
			return false
		}
	}
	if notRoles != nil {
		for _, r := range notRoles {
			if roleMap[r] {
				return false
			}
		}
	}
	return true
}

func childrenOfRoles(n nodes.Node, andRoles []role.Role, notRoles []role.Role) []nodes.Node {
	var children []nodes.Node
	it, _ := tools.Filter(n, "/*")
	for it.Next() {
		child := it.Node().(nodes.Node)
		if containsRoles(child, andRoles, notRoles) {
			children = append(children, child)
		}
	}
	return children
}

func deepChildrenOfRoles(n nodes.Node, andRoles []role.Role, notRoles []role.Role) []nodes.Node {
	var childList []nodes.Node
	it, _ := tools.Filter(n, "/*")
	for it.Next() {
		child := it.Node().(nodes.Node)
		if containsRoles(child, andRoles, notRoles) {
			childList = append(childList, child)
		}
		childList = append(childList, deepChildrenOfRoles(child, andRoles, notRoles)...)
	}
	return childList
}

func countChildrenOfRoles(n nodes.Node, andRoles []role.Role, notRoles []role.Role) int {
	count := 0
	it, _ := tools.Filter(n, "./*")
	for it.Next() {
		child := it.Node().(nodes.Node)
		if containsRoles(child, andRoles, notRoles) {
			count++
		}
	}
	return count
}

func deepCountChildrenOfRoles(n nodes.Node, andRoles []role.Role, notRoles []role.Role) int {
	count := 0
	it, _ := tools.Filter(n, "/*")
	for it.Next() {
		child := it.Node().(nodes.Node)
		if containsRoles(child, andRoles, notRoles) {
			count++
		}
		count += deepCountChildrenOfRoles(child, andRoles, notRoles)
	}
	return count
}
