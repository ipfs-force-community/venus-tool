package cil

import (
	"github.com/ipfs-force-community/venus-tool/client"
	"github.com/urfave/cli/v2"
)

var FlagServer = &cli.StringFlag{
	Name:  "server-addr",
	Usage: "Specify the server address to connect when using cli",
	Value: "127.0.0.1:12580",
}

func getAPI(ctx *cli.Context) (*client.Client, error) {

	serverAddr := "http://localhost:12580"
	if ctx.IsSet(FlagServer.Name) {
		serverAddr = "http://" + ctx.String(FlagServer.Name)
	}

	cli, err := client.New(serverAddr)
	if err != nil {
		return nil, err
	}

	cli.SetVersion("/api/v0")
	return cli, nil
}
