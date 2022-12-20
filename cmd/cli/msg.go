package cil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/filecoin-project/venus/venus-shared/actors/builtin"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/ipfs/go-cid"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/venus-messager/cli/tablewriter"
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
		sendCmd,
		listCmd,
		replaceCmd,
	},
}

var sendCmd = &cli.Command{
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

		client, err := getAPI(ctx)
		if err != nil {
			return err
		}

		var params service.SendReq
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

		var id string
		err = client.Post(ctx.Context, "/msg/send", params, &id)
		if err != nil {
			return err
		}

		// feedback
		fmt.Printf("send message (id: %s ) success\n", id)
		if ctx.Bool("verbose") {
			res := []*service.MsgResp{}
			err := client.Get(ctx.Context, "/msg/"+id, nil, &res)
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

var listCmd = &cli.Command{
	Name:  "list",
	Usage: "list messages",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    "page-index",
			Usage:   "pagination index, start from 1",
			Aliases: []string{"i", "index"},
			Value:   1,
		},
		&cli.IntFlag{
			Name:    "page-size",
			Usage:   "pagination size, default tob 100",
			Aliases: []string{"n"},
			Value:   100,
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
		&cli.StringSliceFlag{
			Name:    "from",
			Usage:   "Specify the sender address",
			Aliases: []string{"f"},
		},
		&cli.BoolFlag{
			Name:  "json",
			Usage: "output in json format",
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
		client, err := getAPI(ctx)
		if err != nil {
			return err
		}

		parseParams := func() (service.QueryMsgReq, error) {

			params := service.QueryMsgReq{}
			nilParams := service.QueryMsgReq{}

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
			params.PageIndex = ctx.Int("page-index")
			params.PageSize = ctx.Int("page-size")

			return params, nil
		}

		params, err := parseParams()
		if err != nil {
			return err
		}

		var res []*service.MsgResp

		err = client.Get(ctx.Context, "/msg/query", params, &res)
		if err != nil {
			return err
		}

		if ctx.Bool("json") {
			return outputWithJson(res)
		}

		return outputWithTable(res, ctx.Bool("verbose"))
	},
}

var replaceCmd = &cli.Command{
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
		client, err := getAPI(cctx)
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
			msgs := []service.MsgResp{}
			params := map[string]interface{}{
				"Nonce": n,
			}

			if cctx.IsSet(flagFrom.Name) {
				f, err := address.NewFromString(cctx.Args().Get(0))
				if err != nil {
					return err
				}
				params["From"] = []address.Address{f}
			}

			err = client.Get(cctx.Context, "/msg/query", params, &msgs)
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
		var cid cid.Cid
		err = client.Post(cctx.Context, "/msg/replace", params, &cid)
		if err != nil {
			return err
		}

		fmt.Println("new message cid: ", cid)
		return nil
	},
}

func outputWithJson(msgs []*service.MsgResp) error {
	bytes, err := json.MarshalIndent(msgs, " ", "\t")
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))
	return nil
}

func outputWithTable(msgs []*service.MsgResp, verbose bool) error {
	var tw = tablewriter.New(
		tablewriter.Col("ID"),
		tablewriter.Col("To"),
		tablewriter.Col("From"),
		tablewriter.Col("Nonce"),
		tablewriter.Col("Value"),
		tablewriter.Col("GasLimit"),
		tablewriter.Col("GasFeeCap"),
		tablewriter.Col("GasPremium"),
		tablewriter.Col("Method"),
		tablewriter.Col("State"),
		tablewriter.Col("ExitCode"),
		tablewriter.Col("CreateAt"),
	)

	for _, msgT := range msgs {
		msg := transformMessage(msgT)
		val := venusTypes.MustParseFIL(msg.Msg.Value.String() + "attofil").String()
		row := map[string]interface{}{
			"ID":         msg.ID,
			"To":         msg.Msg.To,
			"From":       msg.Msg.From,
			"Nonce":      msg.Msg.Nonce,
			"Value":      val,
			"GasLimit":   msg.Msg.GasLimit,
			"GasFeeCap":  msg.Msg.GasFeeCap,
			"GasPremium": msg.Msg.GasPremium,
			"Method":     msg.Msg.Method,
			"State":      msg.State,
			"ErrorMsg":   msgT.ErrorMsg,
			"CreateAt":   msg.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if !verbose {
			if from := msg.Msg.From.String(); len(from) > 9 {
				row["From"] = from[:9] + "..."
			}
			if to := msg.Msg.To.String(); len(to) > 9 {
				row["To"] = to[:9] + "..."
			}
			if len(msg.ID) > 36 {
				row["ID"] = msg.ID[:36] + "..."
			}
			if len(val) > 6 {
				row["Value"] = val[:6] + "..."
			}
		}
		if msg.Receipt != nil {
			row["ExitCode"] = msg.Receipt.ExitCode
		}
		tw.Write(row)
	}

	buf := new(bytes.Buffer)
	if err := tw.Flush(buf); err != nil {
		return err
	}
	fmt.Println(buf)
	return nil
}

type msgTmp struct {
	Version    uint64
	To         address.Address
	From       address.Address
	Nonce      uint64
	Value      abi.TokenAmount
	GasLimit   int64
	GasFeeCap  abi.TokenAmount
	GasPremium abi.TokenAmount
	Method     string
	Params     []byte
}

type receipt struct {
	ExitCode exitcode.ExitCode
	Return   string
	GasUsed  int64
}

type message struct {
	ID string

	UnsignedCid *cid.Cid
	SignedCid   *cid.Cid
	Msg         msgTmp
	Signature   *crypto.Signature

	Height     int64
	Confidence int64
	Receipt    *receipt
	TipSetKey  venusTypes.TipSetKey

	Meta *types.SendSpec

	WalletName string

	State string

	UpdatedAt time.Time
	CreatedAt time.Time
}

func transformMessage(msg *service.MsgResp) *message {
	if msg == nil {
		return nil
	}

	m := &message{
		ID:          msg.ID,
		UnsignedCid: msg.UnsignedCid,
		SignedCid:   msg.SignedCid,
		Signature:   msg.Signature,
		Height:      msg.Height,
		Confidence:  msg.Confidence,
		TipSetKey:   msg.TipSetKey,
		Meta:        msg.Meta,
		WalletName:  msg.WalletName,
		State:       msg.State.String(),

		UpdatedAt: msg.UpdatedAt,
		CreatedAt: msg.CreatedAt,
	}
	if msg.Receipt != nil {
		m.Receipt = &receipt{
			ExitCode: msg.Receipt.ExitCode,
			Return:   string(msg.Receipt.Return),
			GasUsed:  msg.Receipt.GasUsed,
		}
	}

	m.Msg = msgTmp{
		Version:    msg.Version,
		To:         msg.To,
		From:       msg.From,
		Nonce:      msg.Nonce,
		Value:      msg.Value,
		GasLimit:   msg.GasLimit,
		GasFeeCap:  msg.GasFeeCap,
		GasPremium: msg.GasPremium,
		Method:     fmt.Sprint(msg.Message.Method),
		Params:     msg.Params,
	}
	if msg.MethodName != "" {
		m.Msg.Method = msg.MethodName
	}

	return m
}
