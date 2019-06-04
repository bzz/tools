package main

import (
	"context"
	"io/ioutil"
	"path/filepath"

	"github.com/bblfsh/tools"

	"github.com/Sirupsen/logrus"
	"github.com/bblfsh/go-client/v4"
)

type Common struct {
	Address  string `long:"address" description:"server adress to connect to" default:"localhost:9432"`
	Language string `long:"language" description:"language of the input" default:""`
	Args     struct {
		File string `positional-arg-name:"file" required:"true"`
	} `positional-args:"yes"`
}

func (c *Common) execute(args []string, tool tools.Tooler) error {
	logrus.Debugf("executing command")

	ctx := context.Background()
	client, err := bblfsh.NewClientContext(ctx, c.Address)
	if err != nil {
		return err
	}

	logrus.Debugf("reading file %s", c.Args.File)
	content, err := ioutil.ReadFile(c.Args.File)
	if err != nil {
		return err
	}

	uast, _, err := client.NewParseRequest().Context(ctx).
		Mode(bblfsh.Annotated).
		Content(string(content)).
		Language(c.Language).
		Filename(filepath.Base(c.Args.File)).
		UAST()
	if err != nil {
		panic(err)
	}

	return tool.Exec(uast)
}
