package tools

import (
	"io/ioutil"
	"testing"

	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/role"
	"github.com/bblfsh/sdk/v3/uast/uastyaml"
	"github.com/stretchr/testify/require"
)

var (
	n1 = nodes.Object{
		uast.KeyType: nodes.String("module"),
		"body": nodes.Array{
			nodes.Object{
				uast.KeyType:  nodes.String("Statement"),
				uast.KeyRoles: uast.RoleList(role.Statement),
			},
			nodes.Object{
				uast.KeyType:  nodes.String("Statement"),
				uast.KeyRoles: uast.RoleList(role.Statement),
			},
			nodes.Object{
				uast.KeyType:  nodes.String("If"),
				uast.KeyRoles: uast.RoleList(role.Statement, role.If),
			},
		},
	}

	n2 = nodes.Object{
		uast.KeyType: nodes.String("module"),
		"body": nodes.Array{
			nodes.Object{
				uast.KeyType:  nodes.String("Statement"),
				uast.KeyRoles: uast.RoleList(role.Statement),
				"body": nodes.Array{
					nodes.Object{
						uast.KeyType:  nodes.String("Statement"),
						uast.KeyRoles: uast.RoleList(role.Statement),
						"body": nodes.Array{
							nodes.Object{
								uast.KeyType:  nodes.String("If"),
								uast.KeyRoles: uast.RoleList(role.Statement, role.If),
							},
							nodes.Object{
								uast.KeyType:  nodes.String("Statement"),
								uast.KeyRoles: uast.RoleList(role.Statement),
							},
						},
					},
				},
			},
		},
	}
)

func TestCountChildrenOfRole(t *testing.T) {
	require := require.New(t)

	result := countChildrenOfRoles(n1, []role.Role{role.Statement}, nil)
	expect := 3
	require.Equal(expect, result)

	result = countChildrenOfRoles(n2, []role.Role{role.Statement}, nil)
	expect = 1
	require.Equal(expect, result)

	result = deepCountChildrenOfRoles(n1, []role.Role{role.Statement}, nil)
	expect = 3
	require.Equal(expect, result)

	result = deepCountChildrenOfRoles(n2, []role.Role{role.Statement}, nil)
	expect = 4
	require.Equal(expect, result)
}

func TestChildrenOfRole(t *testing.T) {
	require := require.New(t)

	result := childrenOfRoles(n1, []role.Role{role.Statement}, nil)
	expect := 2
	require.Equal(expect, len(result))

	result = childrenOfRoles(n2, []role.Role{role.Statement}, nil)
	expect = 1
	require.Equal(expect, len(result))

	result = deepChildrenOfRoles(n1, []role.Role{role.Statement}, nil)
	expect = 2
	require.Equal(expect, len(result))

	result = deepChildrenOfRoles(n2, []role.Role{role.Statement}, nil)
	expect = 3
	require.Equal(expect, len(result))
}

func TestContainsRole(t *testing.T) {
	require := require.New(t)
	n := nodes.Object{uast.KeyType: nodes.String("module"), uast.KeyRoles: uast.RoleList(role.Statement, role.If)}

	result := containsRoles(n, []role.Role{role.If}, nil)
	require.Equal(true, result)

	result = containsRoles(n, []role.Role{role.Switch}, nil)
	require.Equal(false, result)
}

// func TestExpresionComplex(t *testing.T) {
// 	require := require.New(t)

// 	n := &uast.Node{InternalType: "ifCondition", Roles: []role.Role{uast.If, uast.Condition}, Children: []*uast.Node{
// 		{InternalType: "bool_and", Roles: []role.Role{uast.Operator, uast.Boolean, uast.And}},
// 		{InternalType: "bool_xor", Roles: []role.Role{uast.Operator, uast.Boolean, uast.Xor}},
// 	}}
// 	n2 := &uast.Node{InternalType: "ifCondition", Roles: []role.Role{uast.If, uast.Condition}, Children: []*uast.Node{
// 		{InternalType: "bool_and", Roles: []role.Role{uast.Operator, uast.Boolean, uast.And}, Children: []*uast.Node{
// 			{InternalType: "bool_or", Roles: []role.Role{uast.Operator, uast.Boolean, uast.Or}, Children: []*uast.Node{
// 				{InternalType: "bool_xor", Roles: []role.Role{uast.Operator, uast.Boolean, uast.Xor}},
// 			}},
// 		}},
// 	}}

// 	result := expressionComp(n)
// 	expect := 2
// 	require.Equal(expect, result)

// 	result = expressionComp(n2)
// 	expect = 3
// 	require.Equal(expect, result)
// }

func TestZeroFunction(t *testing.T) {
	require := require.New(t)
	// Empty tree
	n := nodes.Object{uast.KeyType: nodes.String("module")}
	comp := NPathComplexity(n)
	require.Equal(0, len(comp))
}

func TestRealUAST(t *testing.T) {
	fileNames := []string{
		"fixtures/npath/ifelse.java.uast.json",
		"fixtures/npath/do_while.java.uast.json",
		"fixtures/npath/while.java.uast.json",
		"fixtures/npath/for.java.uast.json",
		"fixtures/npath/someFuncs.java.uast.json",
		"fixtures/npath/switch.java.uast.json",
	}

	require := require.New(t)
	var result []int
	for _, name := range fileNames {
		data, err := ioutil.ReadFile(name)
		require.NoError(err)

		ast, err := uastyaml.Unmarshal(data)
		require.NoError(err)

		npathData := NPathComplexity(ast)
		for _, v := range npathData {
			result = append(result, v.Complexity)
		}
	}

	expect := []int{2, 2, 2, 2, 2, 6, 2, 6, 3, 5, 4}

	require.Equal(expect, result)

}
