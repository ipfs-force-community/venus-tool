package cil

import "github.com/urfave/cli/v2"

var FlagServer = &cli.StringFlag{
	Name:  "server",
	Usage: "Specify the server address to connect when using cli",
	Value: "127.0.0.1:12580",
}
