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
	Aliases:   []string{"address"},
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
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		addr := address.Undef
		switch cctx.NArg() {
		case 0:
		case 1:
			addr, err = address.NewFromString(cctx.Args().First())
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("too many arguments")
		}

		addrs, err := api.AddrList(cctx.Context)
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
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if !cctx.Args().Present() {
			return fmt.Errorf("must pass address")
		}

		addr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		params := &service.AddrsOperateReq{
			AddressSpec: msgTypes.AddressSpec{
				Address: addr,
			},
			Operate: service.DeleteAddress,
		}

		err = api.AddrOperate(cctx.Context, params)
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
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if !cctx.Args().Present() {
			return fmt.Errorf("must pass address")
		}

		addr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		params := &service.AddrsOperateReq{
			AddressSpec: msgTypes.AddressSpec{
				Address: addr,
			},
			Operate: service.ForbiddenAddress,
		}

		err = api.AddrOperate(cctx.Context, params)
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
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if !cctx.Args().Present() {
			return fmt.Errorf("must pass address")
		}

		addr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		params := &service.AddrsOperateReq{
			AddressSpec: msgTypes.AddressSpec{
				Address: addr,
			},
			Operate: service.ActiveAddress,
		}

		err = api.AddrOperate(cctx.Context, params)
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
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if !cctx.Args().Present() {
			return fmt.Errorf("must pass address")
		}

		addr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		params := &service.AddrsOperateReq{
			AddressSpec: msgTypes.AddressSpec{
				Address: addr,
			},
			Operate: service.SetAddress,
		}

		isSetSpec := cctx.IsSet("gas-overestimation") || cctx.IsSet("gas-feecap") || cctx.IsSet("max-fee") || cctx.IsSet("base-fee") || cctx.IsSet("gas-over-premium")

		if isSetSpec {
			params.IsSetSpec = isSetSpec
			if cctx.IsSet(flagGasOverPremium.Name) {
				params.GasOverPremium = cctx.Float64(flagGasOverPremium.Name)
			}
			if cctx.IsSet("gas-overestimation") {
				params.GasOverEstimation = cctx.Float64("gas-overestimation")
			}
			if cctx.IsSet("gas-feecap") {
				params.GasFeeCapStr = cctx.String("gas-feecap")
			}
			if cctx.IsSet("max-fee") {
				params.MaxFeeStr = cctx.String("max-fee")
			}
			if cctx.IsSet("base-fee") {
				params.BaseFeeStr = cctx.String("base-fee")
			}
		}

		if cctx.IsSet("num") {
			params.SelectMsgNum = cctx.Uint64("num")
		} else {
			if !isSetSpec {
				return fmt.Errorf("must indicate something to set")
			}
		}

		err = api.AddrOperate(cctx.Context, params)
		if err != nil {
			return err
		}

		fmt.Println("set address success!")
		return nil
	},
}
