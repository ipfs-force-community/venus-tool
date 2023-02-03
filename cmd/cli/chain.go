package cli

import (
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin"
	"github.com/filecoin-project/venus/venus-shared/types"
	_ "github.com/filecoin-project/venus/venus-shared/utils"
	"github.com/ipfs-force-community/venus-tool/utils"
	"github.com/urfave/cli/v2"
)

var ChainCmd = &cli.Command{
	Name:  "chain",
	Usage: "get chain info",
	Subcommands: []*cli.Command{
		chainHeadCmd,
		chainGetActorCmd,
	},
}

var chainHeadCmd = &cli.Command{
	Name:  "head",
	Usage: "get chain head",
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		head, err := api.ChainGetHead(cctx.Context)
		if err != nil {
			return err
		}

		return printJSON(head)
	},
}

var chainGetActorCmd = &cli.Command{
	Name:      "get-actor",
	Usage:     "get actor info",
	ArgsUsage: "<address>",
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() != 1 {
			return fmt.Errorf("must pass exactly one address")
		}

		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		networkName, err := api.ChainGetNetworkName(cctx.Context)
		if err != nil {
			return err
		}

		err = utils.LoadBuiltinActors(cctx.Context, networkName)
		if err != nil {
			return err
		}

		addr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		actor, err := api.ChainGetActor(cctx.Context, addr)
		if err != nil {
			return err
		}

		typeStr := builtin.ActorNameByCode(actor.Code)

		actorInfo := struct {
			Address string
			Balance string
			Nonce   uint64
			Code    string
			Head    string
		}{
			Address: addr.String(),
			Balance: fmt.Sprintf("%s", types.FIL(actor.Balance)),
			Nonce:   actor.Nonce,
			Code:    fmt.Sprintf("%s (%s)", actor.Code, typeStr),
			Head:    actor.Head.String(),
		}

		return printJSON(actorInfo)
	},
}
