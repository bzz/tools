package tools

import "github.com/bblfsh/sdk/v3/uast/nodes"

// Dummy is a sub-command that does not do anything but connecting to bblfshd.
type Dummy struct{}

func (d Dummy) Exec(nodes.Node) error {
	println("It works! You can now proceed with another tool :)")
	return nil
}
