package main

import (
	"fmt"
	"os"

	"github.com/sebnyberg/police-feed-se/cmd/server"
	"github.com/sebnyberg/police-feed-se/cmd/subscribe"
	"github.com/urfave/cli/v2"
)

var pkgName = "police"

func main() {
	app := &cli.App{
		Name:     pkgName,
		HelpName: pkgName,
		Usage:    "Police Event Feed",
		Commands: []*cli.Command{
			server.NewServerCmd(),
			subscribe.NewSubscribeCmd(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
