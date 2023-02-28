package cli

import (
	"fmt"
	"time"

	"github.com/filecoin-project/go-address"
	wallet "github.com/filecoin-project/venus-wallet/storage/wallet"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/urfave/cli/v2"
)

var WalletCmd = &cli.Command{
	Name:    "wallet",
	Usage:   "operate the wallet which is used to send message",
	Aliases: []string{"w"},
	Subcommands: []*cli.Command{
		walletQuerySignRecordCmd,
	},
}

var walletQuerySignRecordCmd = &cli.Command{
	Name:  "sign-record",
	Usage: "query sign record",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "address",
			Usage: "address to query",
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "sign type to query",
		},
		&cli.TimestampFlag{
			Name:     "from",
			Aliases:  []string{"after", "f"},
			Usage:    "from time to query",
			Timezone: time.Local,
			Layout:   "2006-1-2-15:04:05",
		},
		&cli.TimestampFlag{
			Name:     "to",
			Aliases:  []string{"before"},
			Timezone: time.Local,
			Usage:    "to time to query",
			Layout:   "2006-1-2-15:04:05",
		},
		&cli.IntFlag{
			Name:  "limit",
			Usage: "limit to query",
		},
		&cli.IntFlag{
			Name:    "offset",
			Aliases: []string{"skip"},
			Usage:   "offset to query",
		},
		&cli.BoolFlag{
			Name:  "error",
			Usage: "query error record",
		},
		&cli.StringFlag{
			Name:  "id",
			Usage: "query record by id",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Usage:   "verbose output",
			Aliases: []string{"v"},
		},
	},
	Action: func(cctx *cli.Context) error {
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}
		ctx := cctx.Context

		QueryParams := service.WalletSignRecordQueryReq{}

		if cctx.IsSet("address") {
			addrStr := cctx.String("address")
			addr, err := address.NewFromString(addrStr)
			if err != nil {
				return fmt.Errorf("parse address %s : %w", addrStr, err)
			}
			QueryParams.Signer = addr
		}

		if cctx.IsSet("type") {
			t := types.MsgType(cctx.String("type"))
			_, ok := wallet.SupportedMsgTypes[t]
			if !ok {

				fmt.Println("supported types:")
				for k := range wallet.SupportedMsgTypes {
					fmt.Println(k)
				}
				return fmt.Errorf("unsupported type %s", t)
			}
			QueryParams.Type = t
		}

		if cctx.IsSet("from") {
			from := cctx.Timestamp("from")
			QueryParams.After = *from
		}
		if cctx.IsSet("to") {
			to := cctx.Timestamp("to")
			QueryParams.Before = *to
		}
		if cctx.IsSet("limit") {
			limit := cctx.Int("limit")
			QueryParams.Limit = limit
		}
		if cctx.IsSet("offset") {
			offset := cctx.Int("offset")
			QueryParams.Skip = offset
		}
		if cctx.IsSet("error") {
			QueryParams.IsError = cctx.Bool("error")
		}
		if cctx.IsSet("id") {
			QueryParams.ID = cctx.String("id")
		}

		records, err := api.WalletSignRecordQuery(ctx, &QueryParams)
		if err != nil {
			return fmt.Errorf("query sign record: %w", err)
		}

		return printJSON(records)
	},
}
