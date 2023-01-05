package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/filecoin-project/venus/venus-shared/actors/builtin"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/messager"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var (
	flagFrom = &cli.StringFlag{
		Name:  "from",
		Usage: "Specify the sender address",
	}
	flagGasOverPremium = &cli.Float64Flag{
		Name:  "gas-over-premium",
		Usage: "",
	}
	flagVerbose = &cli.BoolFlag{
		Name:    "verbose",
		Usage:   "verbose",
		Aliases: []string{"v"},
	}
)

var MsgCmd = &cli.Command{
	Name:  "msg",
	Usage: "Message related commands",
	Subcommands: []*cli.Command{
		msgDendCmd,
		msgListCmd,
		msgReplaceCmd,
	},
}

var msgDendCmd = &cli.Command{
	Name:      "send",
	Usage:     "Send a message",
	ArgsUsage: "[targetAddress] [amount]",
	Flags: []cli.Flag{
		flagFrom,
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
		flagVerbose,
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() != 2 {
			return fmt.Errorf("'send' expects two arguments, target and amount")
		}

		api := getAPI(ctx)

		var err error
		var params service.MsgSendReq
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

		id, err := api.MsgSend(ctx.Context, &params)
		if err != nil {
			return err
		}

		// feedback
		fmt.Printf("send message (id: %s ) success\n", id)
		if ctx.Bool("verbose") {
			res, err := api.MsgQuery(ctx.Context, &service.MsgQueryReq{ID: id})
			if err != nil {
				return err
			}
			if len(res) == 0 {
				return fmt.Errorf("message not found")
			}
			return outputWithJson(res)
		}

		return nil
	},
}

var msgListCmd = &cli.Command{
	Name:  "list",
	Usage: "list messages",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "from",
			Usage:   "Specify the sender address",
			Aliases: []string{"f"},
		},
		&cli.StringFlag{
			Name:  "id",
			Usage: "Specify the message id",
		},
		&cli.Uint64Flag{
			Name:    "nonce",
			Usage:   "Specify the message nonce",
			Aliases: []string{"n"},
		},
		&cli.BoolFlag{
			Name:  "blocked",
			Usage: "show blocked messages",
		},
		&cli.BoolFlag{
			Name:  "failed",
			Usage: "show failed messages",
		},
		&cli.StringFlag{
			Name:    "time",
			Usage:   "exceeding residence time of blocked msg. Is valid only when [--blocked] flag is set. eg. 3s,3m,3h (default 3h)",
			Aliases: []string{"t"},
			Value:   "3h",
		},
		&cli.BoolFlag{
			Name:  "json",
			Usage: "output in json format",
		},
		&cli.IntFlag{
			Name:    "page-index",
			Usage:   "pagination index, start from 1",
			Aliases: []string{"i", "index"},
			Value:   1,
		},
		&cli.IntFlag{
			Name:  "page-size",
			Usage: "pagination size, default tob 100",
			Value: 100,
		},
		&cli.IntFlag{
			Name:  "state",
			Value: int(types.UnFillMsg),
			Usage: `filter by message state,
state:
  1:  UnFillMsg
  2:  FillMsg
  3:  OnChainMsg
  4:  FailedMsg
  5:  ReplacedMsg
  6:  NoWalletMsg

if [--failed] or [--blocked] is set, [--state] will be ignored 
`,
		},
		flagVerbose,
	},
	Action: func(ctx *cli.Context) error {
		api := getAPI(ctx)
		// client, err := getClient(ctx)
		// if err != nil {
		// 	return err
		// }

		parseParams := func() (service.MsgQueryReq, error) {

			params := service.MsgQueryReq{}
			nilParams := service.MsgQueryReq{}

			if ctx.IsSet("id") {
				params.ID = ctx.String("id")
				return params, nil
			}

			if ctx.IsSet("failed") {
				params.IsFailed = true
				return params, nil
			}

			froms := ctx.StringSlice("from")
			if len(froms) > 0 {
				params.From = make([]address.Address, 0, len(froms))
				for _, from := range froms {
					addr, err := address.NewFromString(from)
					if err != nil {
						return nilParams, fmt.Errorf("failed to parse from address: %w", err)
					}
					params.From = append(params.From, addr)
				}
			}

			if ctx.IsSet("nonce") {
				params.Nonce = ctx.Uint64("nonce")
				if len(params.From) == 0 {
					return nilParams, fmt.Errorf("nonce is set, but from is not set")
				}
				return params, nil
			}

			if ctx.IsSet("blocked") {
				if !ctx.IsSet("time") {
					return nilParams, fmt.Errorf("please set [--time] when [--blocked] is set")
				}
				dur, err := time.ParseDuration(ctx.String("time"))
				if err != nil {
					return nilParams, err
				}
				params.BlockedTime = dur
				params.IsBlocked = true

				return params, nil
			}

			params.State = []types.MessageState{types.MessageState(ctx.Int("state"))}

			if ctx.IsSet("page-index") || ctx.IsSet("page-size") {
				params.PageIndex = ctx.Int("page-index")
				params.PageSize = ctx.Int("page-size")
			}

			return params, nil
		}

		params, err := parseParams()
		if err != nil {
			return err
		}

		res, err := api.MsgQuery(ctx.Context, &params)

		if err != nil {
			return err
		}

		if ctx.Bool("json") {
			return outputWithJson(res)
		}

		return outputMsgWithTable(res, ctx.Bool("verbose"))
	},
}

var msgReplaceCmd = &cli.Command{
	Name:  "replace",
	Usage: "replace a message",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "gas-feecap",
			Usage: "gas feecap for new message (burn and pay to miner, attoFIL/GasUnit)",
		},
		&cli.StringFlag{
			Name:  "gas-premium",
			Usage: "gas price for new message (pay to miner, attoFIL/GasUnit)",
		},
		&cli.Int64Flag{
			Name:  "gas-limit",
			Usage: "gas limit for new message (GasUnit)",
		},
		&cli.BoolFlag{
			Name:  "auto",
			Usage: "automatically reprice the specified message",
		},
		&cli.StringFlag{
			Name:  "max-fee",
			Usage: "Spend up to X attoFIL for this message (applicable for auto mode)",
		},
		&cli.BoolFlag{
			Name:    "nonce",
			Usage:   "use nonce to specify message",
			Aliases: []string{"n"},
		},
		flagFrom,
		flagGasOverPremium,
	},
	ArgsUsage: "<from nonce> | <id>",
	Action: func(cctx *cli.Context) error {
		api := getAPI(cctx)

		if cctx.NArg() != 1 {
			return fmt.Errorf("must specify message id or from nonce")
		}

		var id string
		if cctx.Bool("nonce") {

			n, err := strconv.ParseUint(cctx.Args().Get(0), 10, 64)
			if err != nil {
				return err
			}

			params := service.MsgQueryReq{
				Nonce: n,
			}

			if cctx.IsSet(flagFrom.Name) {
				f, err := address.NewFromString(cctx.Args().Get(0))
				if err != nil {
					return err
				}
				params.From = []address.Address{f}
			}
			msgs, err := api.MsgQuery(cctx.Context, &params)
			if err != nil {
				return fmt.Errorf("could not find referenced message: %w", err)
			}
			msg := msgs[0]
			id = msg.ID
		} else {
			id = cctx.Args().First()
		}

		parseParams := func() (*messager.ReplacMessageParams, error) {
			params := messager.ReplacMessageParams{
				Auto:           cctx.Bool("auto"),
				GasLimit:       cctx.Int64("gas-limit"),
				GasOverPremium: cctx.Float64(flagGasOverPremium.Name),
			}

			if cctx.IsSet("max-fee") {
				maxFee, err := venusTypes.ParseFIL(cctx.String("max-fee"))
				if err != nil {
					return nil, fmt.Errorf("parse max fee failed: %v", err)
				}
				params.MaxFee = big.Int(maxFee)
			}
			if cctx.IsSet("gas-premium") {
				gasPremium, err := venusTypes.BigFromString(cctx.String("gas-premium"))
				if err != nil {
					return nil, fmt.Errorf("parse gas premium failed: %v", err)
				}
				params.GasPremium = gasPremium
			}
			if cctx.IsSet("gas-feecap") {
				gasFeecap, err := venusTypes.BigFromString(cctx.String("gas-feecap"))
				if err != nil {
					return nil, fmt.Errorf("parse gas feecap failed: %v", err)
				}
				params.GasFeecap = gasFeecap
			}

			return &params, nil
		}

		params, err := parseParams()
		if err != nil {
			return err
		}
		params.ID = id
		cid, err := api.MsgReplace(cctx.Context, params)
		if err != nil {
			return err
		}

		fmt.Println("new message cid: ", cid)
		return nil
	},
}
