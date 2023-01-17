package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/urfave/cli/v2"
)

var MultiSigCmd = &cli.Command{
	Name:  "msig",
	Usage: "Manage multisig wallets",
	Subcommands: []*cli.Command{
		multisigCreateCmd,
		multisigProposeCmd,
		multisigProposeListCmd,
		multisigApproveCmd,
		multisigAddSignerCmd,
	},
}

var multisigCreateCmd = &cli.Command{
	Name:      "create",
	Usage:     "Create a multisig wallet",
	ArgsUsage: "<signer address>...",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:  "required",
			Usage: "number of required approvals (uses number of signers provided if omitted)",
		},
		&cli.StringFlag{
			Name:  "value",
			Usage: "initial funds to give to multisig",
			Value: "0",
		},
		&cli.Int64Flag{
			Name:  "duration",
			Usage: "length of the period over which funds unlock",
			Value: 0,
		},
		&cli.StringFlag{
			Name:  "from",
			Usage: "account to send the create message from (uses the first signer if omitted)",
		},
	},
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() < 1 {
			return fmt.Errorf("must specify at least one signer")
		}

		var signers []address.Address
		for i := 0; i < cctx.NArg(); i++ {
			addr, err := address.NewFromString(cctx.Args().Get(i))
			if err != nil {
				return err
			}
			signers = append(signers, addr)
		}

		var from address.Address
		if cctx.IsSet("from") {
			from, err = address.NewFromString(cctx.String("from"))
			if err != nil {
				return err
			}
		} else {
			from = signers[0]
		}

		var required uint64
		if cctx.IsSet("required") {
			required = cctx.Uint64("required")
		} else {
			required = uint64(len(signers))
		}

		value, err := types.ParseFIL(cctx.String("value"))
		if err != nil {
			return err
		}

		duration := abi.ChainEpoch(cctx.Int64("duration"))

		req := &service.MultisigCreateReq{
			Signers:            signers,
			ApprovalsThreshold: required,
			Value:              types.BigInt(value),
			LockedDuration:     duration,
			From:               from,
		}

		newAddr, err := api.MsigCreate(cctx.Context, req)
		if err != nil {
			return err
		}

		fmt.Printf("Created new multisig wallet at address %s \n", newAddr)
		return nil
	},
}

var multisigProposeCmd = &cli.Command{
	Name:      "propose",
	Usage:     "Propose a multisig transaction",
	ArgsUsage: "<multisig address> <proposer address> <destination address> <value>",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:  "method",
			Usage: "specify method to invoke",
			Value: uint64(builtin.MethodSend),
		},
		&cli.StringFlag{
			Name:  "params-json",
			Usage: "specify invocation parameters in json",
		},
	},
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() != 4 {
			return fmt.Errorf("must specify multisig address, proposer address, destination address, and value")
		}

		msigAddr, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		to, err := address.NewFromString(cctx.Args().Get(2))
		if err != nil {
			return err
		}

		value, err := types.ParseFIL(cctx.Args().Get(3))
		if err != nil {
			return err
		}

		method := abi.MethodNum(cctx.Uint64("method"))

		var params service.EncodedParams
		if cctx.IsSet("params-json") && cctx.IsSet("params-hex") {
			return fmt.Errorf("must specify only one of params-json and params-hex")
		}
		if cctx.IsSet("params-json") {
			params.Data, err = json.Marshal(cctx.String("params-json"))
			if err != nil {
				return err
			}
			params.EncType = service.EncJson
		}
		if cctx.IsSet("params-hex") {
			params.Data, err = hex.DecodeString(cctx.String("params-hex"))
			if err != nil {
				return err
			}
			params.EncType = service.EncHex
		}

		req := &service.MultisigProposeReq{
			Msig:   msigAddr,
			From:   from,
			To:     to,
			Value:  types.BigInt(value),
			Method: method,
			Params: params,
		}

		ret, err := api.MsigPropose(cctx.Context, req)
		if err != nil {
			return err
		}

		return printJSON(ret)
	},
}

var multisigProposeListCmd = &cli.Command{
	Name:      "propose-list",
	Usage:     "List pending multisig transactions",
	ArgsUsage: "<multisig address>",
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() != 1 {
			return fmt.Errorf("must specify multisig address only")
		}

		msigAddr, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		ret, err := api.MsigListPropose(cctx.Context, msigAddr)
		if err != nil {
			return err
		}

		type ProposeOutput struct {
			TxID       int64
			To         address.Address
			Value      types.FIL
			RawParams  []byte
			Params     json.Marshaler
			Method     abi.MethodNum
			MethodName string
			Approved   []address.Address
		}

		var out []ProposeOutput
		for _, p := range ret {
			var b []byte
			var err error
			if len(p.Params) > 0 {
				b, err = api.MsgDecodeParam2Json(cctx.Context, &service.MsgDecodeParamReq{
					To:     p.To,
					Method: p.Method,
					Params: p.Params,
				})
				if err != nil {
					return err
				}
			} else {
				b = []byte("null")
			}

			params := marshaler{
				Data: b,
			}

			methodName, err := api.MsgGetMethodName(cctx.Context, &service.MsgGetMethodNameReq{
				To:     p.To,
				Method: p.Method,
			})
			if err != nil {
				return err
			}

			out = append(out, ProposeOutput{
				TxID:       p.ID,
				To:         p.To,
				Value:      types.FIL(p.Value),
				RawParams:  p.Params,
				Params:     &params,
				Method:     p.Method,
				MethodName: methodName,
				Approved:   p.Approved,
			})
		}

		return printJSON(out)
	},
}

var multisigAddSignerCmd = &cli.Command{
	Name:      "add",
	Usage:     "Add a signer to a multisig wallet",
	ArgsUsage: "<multisig address> <proposer address> <signer address>",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "increase-threshold",
			Aliases: []string{"inc"},
			Usage:   "whether increase threshold",
		},
	},
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() <= 3 {
			return fmt.Errorf("must specify multisig address, proposer address, and signer address")
		}

		msigAddr, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		signer, err := address.NewFromString(cctx.Args().Get(2))
		if err != nil {
			return err
		}

		inc := cctx.Bool("increase-threshold")

		req := &service.MultisigAddSignerReq{
			Msig:              msigAddr,
			Proposer:          from,
			NewSigner:         signer,
			IncreaseThresHold: inc,
		}

		ret, err := api.MsigAddSigner(cctx.Context, req)
		if err != nil {
			return err
		}

		return printJSON(ret)
	},
}

var multisigApproveCmd = &cli.Command{
	Name:      "approve",
	Usage:     "Approve a multisig transaction",
	ArgsUsage: "<multisig address> <proposer address> <txid>",
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() != 3 {
			return fmt.Errorf("must specify multisig address, proposer address, and txid")
		}

		msigAddr, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		txid, err := strconv.ParseUint(cctx.Args().Get(2), 10, 64)
		if err != nil {
			return err
		}

		req := &service.MultisigApproveReq{
			Msig:     msigAddr,
			Proposer: from,
			TxID:     txid,
		}

		ret, err := api.MsigApprove(cctx.Context, req)
		if err != nil {
			return err
		}

		return printJSON(ret)
	},
}
