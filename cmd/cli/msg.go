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
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/messager"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
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
	Name:    "msg",
	Usage:   "Message related commands",
	Aliases: []string{"message"},
	Subcommands: []*cli.Command{
		msgSendCmd,
		msgListCmd,
		msgReplaceCmd,
	},
}

var msgSendCmd = &cli.Command{
	Name:      "send",
	Usage:     "Send a message",
	ArgsUsage: "<targetAddress> <amount>",
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
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("'send' expects two arguments, target and amount")
		}

		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		var req service.MsgSendReq
		req.To, err = address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return fmt.Errorf("failed to parse target address: %w", err)
		}

		val, err := types.ParseFIL(cctx.Args().Get(1))
		if err != nil {
			return fmt.Errorf("failed to parse amount: %w", err)
		}
		req.Value = abi.TokenAmount(val)

		addr, err := address.NewFromString(cctx.String("from"))
		if err != nil {
			return fmt.Errorf("failed to parse from address: %w", err)
		}
		req.From = addr

		req.Method = abi.MethodNum(cctx.Uint64("method"))

		gfc, err := types.BigFromString(cctx.String("max-fee"))
		if err != nil {
			return err
		}
		req.MaxFee = gfc

		req.GasOverPremium = cctx.Float64("gas-over-premium")

		req.GasOverEstimation = cctx.Float64("gas-over-estimation")

		if cctx.IsSet("params-json") {
			req.Params.Data = []byte(cctx.String("params-json"))
			req.Params.EncType = service.EncJson
		}
		if cctx.IsSet("params-hex") {
			if len(req.Params.Data) != 0 {
				return fmt.Errorf("can only specify one of 'params-json' and 'params-hex'")
			}
			req.Params.Data = []byte(cctx.String("params-hex"))
			req.Params.EncType = service.EncHex
		}

		id, err := api.MsgSend(cctx.Context, &req)
		if err != nil {
			return err
		}

		// feedback
		fmt.Printf("send message (id: %s ) success\n", id)
		if cctx.Bool("verbose") {
			res, err := api.MsgQuery(cctx.Context, &service.MsgQueryReq{ID: id})
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
			Value: int(msgTypes.UnFillMsg),
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
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		parseParams := func() (service.MsgQueryReq, error) {

			params := service.MsgQueryReq{}
			nilParams := service.MsgQueryReq{}

			if cctx.IsSet("id") {
				params.ID = cctx.String("id")
				return params, nil
			}

			if cctx.IsSet("failed") {
				params.IsFailed = true
				return params, nil
			}

			froms := cctx.StringSlice("from")
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

			if cctx.IsSet("nonce") {
				params.Nonce = cctx.Uint64("nonce")
				if len(params.From) == 0 {
					return nilParams, fmt.Errorf("nonce is set, but from is not set")
				}
				return params, nil
			}

			if cctx.IsSet("blocked") {
				if !cctx.IsSet("time") {
					return nilParams, fmt.Errorf("please set [--time] when [--blocked] is set")
				}
				dur, err := time.ParseDuration(cctx.String("time"))
				if err != nil {
					return nilParams, err
				}
				params.BlockedTime = dur
				params.IsBlocked = true

				return params, nil
			}

			params.State = []msgTypes.MessageState{msgTypes.MessageState(cctx.Int("state"))}

			if cctx.IsSet("page-index") || cctx.IsSet("page-size") {
				params.PageIndex = cctx.Int("page-index")
				params.PageSize = cctx.Int("page-size")
			}

			return params, nil
		}

		params, err := parseParams()
		if err != nil {
			return err
		}

		res, err := api.MsgQuery(cctx.Context, &params)

		if err != nil {
			return err
		}

		if cctx.Bool("json") {
			return outputWithJson(res)
		}

		return outputMsgWithTable(res, cctx.Bool("verbose"))
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
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

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
				maxFee, err := types.ParseFIL(cctx.String("max-fee"))
				if err != nil {
					return nil, fmt.Errorf("parse max fee failed: %v", err)
				}
				params.MaxFee = big.Int(maxFee)
			}
			if cctx.IsSet("gas-premium") {
				gasPremium, err := types.BigFromString(cctx.String("gas-premium"))
				if err != nil {
					return nil, fmt.Errorf("parse gas premium failed: %v", err)
				}
				params.GasPremium = gasPremium
			}
			if cctx.IsSet("gas-feecap") {
				gasFeecap, err := types.BigFromString(cctx.String("gas-feecap"))
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
