package cil

import (
	"fmt"

	"github.com/filecoin-project/venus/venus-shared/actors/builtin"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
)

var MsgCmd = &cli.Command{
	Name:  "msg",
	Usage: "Message related commands",
	Subcommands: []*cli.Command{
		MsgSendCmd,
	},
}

var MsgSendCmd = &cli.Command{
	Name:      "send",
	Usage:     "Send a message",
	ArgsUsage: "[targetAddress] [amount]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Usage:    "optionally specify the address to send",
			Required: true,
		},
		&cli.Uint64Flag{
			Name:  "method",
			Usage: "specify method to invoke",
			Value: uint64(builtin.MethodSend),
		},
		&cli.StringFlag{
			Name:  "params-json",
			Usage: "specify invocation parameters in json",
		},
		&cli.StringFlag{
			Name:  "params-hex",
			Usage: "specify invocation parameters in hex",
		},
		&cli.StringFlag{
			Name:  "max-fee",
			Usage: "indicate the max fee can be used to send message in AttoFIL",
			Value: "0",
		},
		&cli.Float64Flag{
			Name:  "gas-over-premium",
			Usage: "the ratio of gas premium base on estimated gas premium",
			Value: 0,
		},

		&cli.Float64Flag{
			Name:  "gas-over-estimation",
			Usage: "the ratio of gas limit base on estimated gas used",
			Value: 0,
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() != 2 {
			return fmt.Errorf("'send' expects two arguments, target and amount")
		}

		client, err := getClient(ctx)
		if err != nil {
			return err
		}

		var params service.SendParams
		params.To, err = address.NewFromString(ctx.Args().Get(0))
		if err != nil {
			return fmt.Errorf("failed to parse target address: %w", err)
		}

		val, err := venusTypes.ParseFIL(ctx.Args().Get(1))
		if err != nil {
			return fmt.Errorf("failed to parse amount: %w", err)
		}
		params.Value = abi.TokenAmount(val)

		addr, err := address.NewFromString(ctx.String("from"))
		if err != nil {
			return fmt.Errorf("failed to parse from address: %w", err)
		}
		params.From = addr

		params.Method = abi.MethodNum(ctx.Uint64("method"))

		gfc, err := venusTypes.BigFromString(ctx.String("max-fee"))
		if err != nil {
			return err
		}
		params.MaxFee = gfc

		params.GasOverPremium = ctx.Float64("gas-over-premium")

		params.GasOverEstimation = ctx.Float64("gas-over-estimation")

		if ctx.IsSet("params-json") {
			params.Params = []byte(ctx.String("params-json"))
			params.EncType = service.EncJson
		}
		if ctx.IsSet("params-hex") {
			if len(params.Params) != 0 {
				return fmt.Errorf("can only specify one of 'params-json' and 'params-hex'")
			}
			params.Params = []byte(ctx.String("params-hex"))
			params.EncType = service.EncHex
		}

		var res string
		err = client.Post(ctx.Context, "/api/v0/send", params, &res)
		if err != nil {
			return err
		}

		// feedback
		fmt.Printf("send message (id: %s ) success\n", res)

		return nil
	},
}
