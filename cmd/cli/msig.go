package cli

import (
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
		multisigInfoCmd,
		multisigProposeCmd,
		multisigProposeListCmd,
		multisigApproveCmd,
		multisigCancelCmd,
		multisigCreateCmd,
		multisigAddSignerCmd,
		multisigRemoveSignerCmd,
		multisigSwapSignerCmd,
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
		&cli.StringFlag{
			Name:  "params-hex",
			Usage: "specify invocation parameters in hex",
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
			params.Data = cctx.String("params-json")
			params.EncType = service.EncJson
		}
		if cctx.IsSet("params-hex") {
			params.Data = cctx.String("params-hex")
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
			Value      types.FIL
			MethodName string
			MethodId   abi.MethodNum
			Params     json.Marshaler
			RawParams  []byte
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
				Value:      types.FIL(p.Value),
				RawParams:  p.Params,
				Params:     json.RawMessage(b),
				MethodId:   p.Method,
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

		req := &service.MultisigChangeSignerReq{
			Msig:           msigAddr,
			Proposer:       from,
			NewSigner:      signer,
			AlterThresHold: inc,
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

		req := &service.MultisigTransactionReq{
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

var multisigCancelCmd = &cli.Command{
	Name:      "cancel",
	Usage:     "Cancel a multisig transaction",
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

		req := &service.MultisigTransactionReq{
			Msig:     msigAddr,
			Proposer: from,
			TxID:     txid,
		}

		err = api.MsigCancel(cctx.Context, req)
		if err != nil {
			return err
		}

		fmt.Printf("Cancelled transaction(%d) successfully \n", txid)
		return nil
	},
}

var multisigInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "Get info about a multisig wallet",
	ArgsUsage: "<multisig address>",
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() != 1 {
			return fmt.Errorf("must specify multisig address")
		}

		msigAddr, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		ret, err := api.MsigInfo(cctx.Context, msigAddr)
		if err != nil {
			return err
		}

		type output struct {
			types.MsigInfo
			InitialBalance types.FIL
			CurrentBalance types.FIL
			LockBalance    types.FIL
		}

		out := output{
			MsigInfo:       *ret,
			InitialBalance: types.FIL(ret.InitialBalance),
			CurrentBalance: types.FIL(ret.CurrentBalance),
			LockBalance:    types.FIL(ret.LockBalance),
		}

		return printJSON(out)
	},
}

var multisigRemoveSignerCmd = &cli.Command{
	Name:      "remove",
	Usage:     "Remove a signer from a multisig wallet",
	ArgsUsage: "<multisig address> <proposer address> <signer address>",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "decrease-threshold",
			Usage:   "decrease the multisig threshold by 1",
			Aliases: []string{"dec"},
		},
	},
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() != 3 {
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

		dec := cctx.Bool("decrease-threshold")

		req := &service.MultisigChangeSignerReq{
			Msig:           msigAddr,
			Proposer:       from,
			NewSigner:      signer,
			AlterThresHold: dec,
		}

		ret, err := api.MsigRemoveSigner(cctx.Context, req)
		if err != nil {
			return err
		}

		return printJSON(ret)
	},
}

var multisigSwapSignerCmd = &cli.Command{
	Name:      "swap",
	Usage:     "Swap a signer in a multisig wallet",
	ArgsUsage: "<multisig address> <proposer address> <old signer address> <new signer address>",
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() != 4 {
			return fmt.Errorf("must specify multisig address, proposer address, old signer address, and new signer address")
		}

		msigAddr, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		oldSigner, err := address.NewFromString(cctx.Args().Get(2))
		if err != nil {
			return err
		}

		newSigner, err := address.NewFromString(cctx.Args().Get(3))
		if err != nil {
			return err
		}

		req := &service.MultisigSwapSignerReq{
			Msig:      msigAddr,
			Proposer:  from,
			OldSigner: oldSigner,
			NewSigner: newSigner,
		}

		ret, err := api.MsigSwapSigner(cctx.Context, req)
		if err != nil {
			return err
		}

		return printJSON(ret)
	},
}
