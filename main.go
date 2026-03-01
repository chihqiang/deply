package main

import (
	"chihqiang/deply/cmdx"
	"chihqiang/deply/flagx"
	"context"
	"os"

	"github.com/chihqiang/logx"
	"github.com/urfave/cli/v3"
)

var (
	version = "main"
)

func main() {
	app := &cli.Command{
		Name:    "deply",
		Usage:   "Push it, roll it, own it",
		Version: version,
		Flags:   flagx.SSHFlags(),
		Commands: []*cli.Command{
			cmdx.Publish(),
			cmdx.History(),
			cmdx.Rollback(),
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		logx.Error("%+v", err)
		os.Exit(1)
	}
}
