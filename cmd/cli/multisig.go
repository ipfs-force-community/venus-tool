package cli

import (
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/urfave/cli/v2"
)

var MultiSigCmd = &cli.Command{
	Name:  "msig",
	Usage: "Manage multisig wallets",
	Subcommands: []*cli.Command{
		multiSigCreateCmd,
	},
}

var multiSigCreateCmd = &cli.Command{
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

		req := &service.MultiSigCreateReq{
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
