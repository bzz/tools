package tools

import (
	"github.com/bblfsh/go-client/v4/tools"
	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
)

// Tokenizer sub-command outputs every token to STDOUT.
type Tokenizer struct{}

func (t Tokenizer) Exec(node nodes.Node) error {
	for _, token := range Tokens(node) {
		print(token)
	}
	return nil
}

// Tokens returns a slice of tokens contained in the node.
func Tokens(n nodes.Node) []string {
	var tokens []string
	iter := tools.NewIterator(n, tools.PreOrder)

	for n := range tools.Iterate(iter) {
		token := uast.TokenOf(n)
		if token != "" {
			tokens = append(tokens, token)
		}
	}
	return tokens
}
