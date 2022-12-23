package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/venus/venus-shared/types/market"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/ipfs/go-cid"
	"github.com/urfave/cli/v2"
)

var DealCmd = &cli.Command{
	Name:  "deal",
	Usage: "Manage deals",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "retrieval",
			Usage: "Manage retrieval deals rather than storage deals which is default",
		},
	},
	Subcommands: []*cli.Command{
		dealListCmd,
		dealUpdateCmd,
	},
}

var dealListCmd = &cli.Command{
	Name:      "list",
	Usage:     "List deals",
	ArgsUsage: "[miner address]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		client, err := getAPI(cctx)
		if err != nil {
			return err
		}

		verbose := cctx.Bool("verbose")
		mAddr := address.Undef
		if cctx.Args().Len() >= 1 {
			mAddr, err = address.NewFromString(cctx.Args().First())
			if err != nil {
				return err
			}
		}

		if cctx.Bool("retrieval") {
			deals := []market.ProviderDealState{}
			err := client.Get(ctx, "/deal/retrieval", nil, &deals)
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)

			_, _ = fmt.Fprintf(w, "Receiver\tDealID\tPayload\tState\tPricePerByte\tBytesSent\tPaied\tInterval\tMessage\n")

			for _, deal := range deals {
				payloadCid := deal.PayloadCID.String()

				_, _ = fmt.Fprintf(w,
					"%s\t%d\t%s\t%s\t%s\t%d\t%d\t%d\t%s\n",
					deal.Receiver.String(),
					deal.ID,
					"..."+payloadCid[len(payloadCid)-8:],
					retrievalmarket.DealStatuses[deal.Status],
					deal.PricePerByte.String(),
					deal.TotalSent,
					deal.FundsReceived,
					deal.CurrentInterval,
					deal.Message,
				)
			}

			return w.Flush()
		}

		deals := []market.MinerDeal{}
		err = client.Get(ctx, "/deal/storage", mAddr, &deals)
		if err != nil {
			return err
		}

		return outputStorageDeals(os.Stdout, deals, verbose)
	},
}

var dealUpdateCmd = &cli.Command{
	Name:  "update",
	Usage: "Update deal",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "proposalcid",
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "really-do-it",
			Usage: "Actually send transaction performing the action",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "piece-state",
			Usage: "Undefine | Assigned | Packing | Proving, empty means no change",
		},
		&cli.StringFlag{
			Name:  "state",
			Usage: dealStateUsage(),
		},
	},
	Action: func(cctx *cli.Context) error {
		client, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.Bool("retrieval") {
			return fmt.Errorf("retrieval deals cannot be updated")
		}

		proposalCid, err := cid.Decode(cctx.String("proposalcid"))
		if err != nil {
			return err
		}
		var isParamOk bool
		var state storagemarket.StorageDealStatus
		var pieceStatus market.PieceStatus

		if cctx.IsSet("state") {
			isParamOk = true
			state = StringToStorageState[cctx.String("state")]
		}

		if cctx.IsSet("piece-state") {
			pieceStatus = market.PieceStatus(cctx.String("piece-state"))
			isParamOk = true
		}

		if !isParamOk {
			return fmt.Errorf("must set 'state' or 'piece-state'")
		}

		if !cctx.Bool("really-do-it") {
			fmt.Println("Pass --really-do-it to actually execute this action")
			return nil
		}

		return client.Post(cctx.Context, "/deal/storage/state", &service.StorageDealUpdateStateReq{
			ProposalCid: proposalCid,
			State:       state,
			PieceStatus: pieceStatus,
		}, nil)
	},
}

var dealStateUsage = func() string {
	const c, spliter = 5, " | "
	size := len(StringToStorageState)
	states := make([]string, 0, size+size/c)
	idx := 0
	for s := range StringToStorageState {
		states = append(states, s)
		idx++
		states = append(states, spliter)
		if idx%c == 0 {
			states = append(states, "\n\t")
			continue
		}
	}

	usage := strings.Join(states, "")
	{
		size := len(usage)
		if size > 3 && usage[size-3:] == spliter {
			usage = usage[:size-3]
		}
	}
	return usage + ", set to 'StorageDealUnknown' means no change"
}
