package cli

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-address"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/urfave/cli/v2"
)

var AddrCmd = &cli.Command{
	Name:      "addr",
	Usage:     "operate the address which is used to send message",
	ArgsUsage: "[address]",
	Subcommands: []*cli.Command{
		addrListCmd,
		addrDeleteCmd,
		addrActiveCmd,
		addrForbiddenCmd,
		addrSetCmd,
	},
}

var addrListCmd = &cli.Command{
	Name:      "list",
	Usage:     "list address",
	ArgsUsage: "[address]",
	Action: func(ctx *cli.Context) error {
		api := getAPI(ctx)

		addr := address.Undef
		var err error
		switch ctx.NArg() {
		case 0:
		case 1:
			addr, err = address.NewFromString(ctx.Args().First())
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("too many arguments")
		}

		addrs, err := api.AddrList(ctx.Context)
		if err != nil {
			return err
		}

		if len(addrs) == 0 {
			return nil
		}

		if addr != address.Undef {
			for _, a := range addrs {
				if a.Addr == addr {
					return printJSON(a)
				}
			}
		} else {
			bytes, err := json.MarshalIndent(addrs, " ", "\t")
			if err != nil {
				return err
			}
			fmt.Println(string(bytes))
		}
		return nil
	},
}

var addrDeleteCmd = &cli.Command{
	Name:      "del",
	Usage:     "delete address",
	ArgsUsage: "<address>",
	Action: func(ctx *cli.Context) error {
		api := getAPI(ctx)

		if !ctx.Args().Present() {
			return fmt.Errorf("must pass address")
		}

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}

		params := &service.AddrsOperateReq{
			AddressSpec: msgTypes.AddressSpec{
				Address: addr,
			},
			Operate: service.DeleteAddress,
		}

		err = api.AddrOperate(ctx.Context, params)
		if err != nil {
			return err
		}

		fmt.Println("delete address success!")
		return nil
	},
}

var addrForbiddenCmd = &cli.Command{
	Name:      "forbidden",
	Usage:     "forbidden address",
	ArgsUsage: "<address>",
	Action: func(ctx *cli.Context) error {
		api := getAPI(ctx)

		if !ctx.Args().Present() {
			return fmt.Errorf("must pass address")
		}

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}

		params := &service.AddrsOperateReq{
			AddressSpec: msgTypes.AddressSpec{
				Address: addr,
			},
			Operate: service.ForbiddenAddress,
		}

		err = api.AddrOperate(ctx.Context, params)
		if err != nil {
			return err
		}
		fmt.Println("forbidden address success!")

		return nil
	},
}

var addrActiveCmd = &cli.Command{
	Name:      "active",
	Usage:     "activate a frozen address",
	ArgsUsage: "<address>",
	Action: func(ctx *cli.Context) error {
		api := getAPI(ctx)

		if !ctx.Args().Present() {
			return fmt.Errorf("must pass address")
		}

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}

		params := &service.AddrsOperateReq{
			AddressSpec: msgTypes.AddressSpec{
				Address: addr,
			},
			Operate: service.ActiveAddress,
		}

		err = api.AddrOperate(ctx.Context, params)
		if err != nil {
			return err
		}
		fmt.Println("active address success!")

		return nil
	},
}

var addrSetCmd = &cli.Command{
	Name:      "set",
	Usage:     "Address setting fee associated configuration",
	ArgsUsage: "<address>",
	Flags: []cli.Flag{
		&cli.Float64Flag{
			Name:  "gas-overestimation",
			Usage: "Estimate the coefficient of gas",
		},
		&cli.StringFlag{
			Name:  "gas-feecap",
			Usage: "Gas feecap for a message (burn and pay to miner, attoFIL/GasUnit)",
		},
		&cli.StringFlag{
			Name:  "max-fee",
			Usage: "Spend up to X attoFIL for message",
		},
		&cli.StringFlag{
			Name:  "base-fee",
			Usage: "",
		},
		&cli.Uint64Flag{
			Name:  "num",
			Usage: "the number of one address selection message",
		},
		flagGasOverPremium,
	},
	Action: func(ctx *cli.Context) error {
		api := getAPI(ctx)

		if !ctx.Args().Present() {
			return fmt.Errorf("must pass address")
		}

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}

		params := &service.AddrsOperateReq{
			AddressSpec: msgTypes.AddressSpec{
				Address: addr,
			},
			Operate: service.SetAddress,
		}

		isSetSpec := ctx.IsSet("gas-overestimation") || ctx.IsSet("gas-feecap") || ctx.IsSet("max-fee") || ctx.IsSet("base-fee") || ctx.IsSet("gas-over-premium")

		if isSetSpec {
			params.IsSetSpec = isSetSpec
			if ctx.IsSet(flagGasOverPremium.Name) {
				params.GasOverPremium = ctx.Float64(flagGasOverPremium.Name)
			}
			if ctx.IsSet("gas-overestimation") {
				params.GasOverEstimation = ctx.Float64("gas-overestimation")
			}
			if ctx.IsSet("gas-feecap") {
				params.GasFeeCapStr = ctx.String("gas-feecap")
			}
			if ctx.IsSet("max-fee") {
				params.MaxFeeStr = ctx.String("max-fee")
			}
			if ctx.IsSet("base-fee") {
				params.BaseFeeStr = ctx.String("base-fee")
			}
		}

		if ctx.IsSet("num") {
			params.SelectMsgNum = ctx.Uint64("num")
		} else {
			if !isSetSpec {
				return fmt.Errorf("must indicate something to set")
			}
		}

		err = api.AddrOperate(ctx.Context, params)
		if err != nil {
			return err
		}

		fmt.Println("set address success!")
		return nil
	},
}
