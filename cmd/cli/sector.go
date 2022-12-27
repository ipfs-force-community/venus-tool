package cli

import (
	"fmt"
	"strconv"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/urfave/cli/v2"
)

var SectorCmd = &cli.Command{
	Name:  "sector",
	Usage: "Interact with sectors",
	Subcommands: []*cli.Command{
		sectorExtendCmd,
	},
}

var sectorExtendCmd = &cli.Command{
	Name:      "extend",
	Usage:     "Extend a sector's lifetime",
	ArgsUsage: "[sectorNumber ...]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "miner",
			Usage:    "miner address",
			Required: true,
		},
		&cli.Int64Flag{
			Name:     "expiration",
			Usage:    "new expiration epoch",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must pass at least one sector number")
		}

		ctx := cctx.Context
		client, err := getAPI(cctx)
		if err != nil {
			return err
		}

		miner, err := address.NewFromString(cctx.String("miner"))
		if err != nil {
			return err
		}

		expiration := cctx.Int64("expiration")

		req := service.SectorExtendReq{
			Miner:      miner,
			Expiration: abi.ChainEpoch(expiration),
		}

		for i, s := range cctx.Args().Slice() {
			id, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return fmt.Errorf("could not parse sector %d: %w", i, err)
			}
			req.SectorNumbers = append(req.SectorNumbers, abi.SectorNumber(id))
		}

		return client.Post(ctx, "/sector/extend", req, nil)
	},
}
