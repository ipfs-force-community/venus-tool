package cli

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/urfave/cli/v2"
)

var MinerCmd = &cli.Command{
	Name:  "miner",
	Usage: "manage miner",
	Subcommands: []*cli.Command{
		minerAskCmd,
	},
}

var minerAskCmd = &cli.Command{
	Name:  "ask",
	Usage: "manage miner ask",
	Subcommands: []*cli.Command{
		minerGetAskCmd,
		minerSetAskCmd,
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "retrieval",
			Usage: "manage retrieval ask rather than storage ask which is default",
		},
	},
}

var minerGetAskCmd = &cli.Command{
	Name:      "get",
	Usage:     "get miner ask",
	ArgsUsage: "[miner address]",
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context

		client, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.Args().Len() != 1 {
			return errors.New("must specify miner address")
		}
		mAddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return fmt.Errorf("para `miner` is invalid: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)

		if cctx.Bool("retrieval") {
			ask := retrievalmarket.Ask{}
			err := client.Get(ctx, "/miner/ask/retrieval", mAddr, &ask)
			if err != nil {
				return err
			}

			fmt.Fprintf(w, "Price per Byte\tUnseal Price\tPayment Interval\tPayment Interval Increase\n")

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				types.FIL(ask.PricePerByte),
				types.FIL(ask.UnsealPrice),
				units.BytesSize(float64(ask.PaymentInterval)),
				units.BytesSize(float64(ask.PaymentIntervalIncrease)),
			)
		} else {

			ask := storagemarket.StorageAsk{}
			err = client.Get(ctx, "/miner/ask/storage", mAddr, &ask)
			if err != nil {
				return err
			}

			fmt.Fprintf(w, "Price per GiB/Epoch\tVerified\tMin. Piece Size (padded)\tMax. Piece Size (padded)\tExpiry (Epoch)\tExpiry (Appx. Rem. Time)\tSeq. No.\n")

			head := types.TipSet{}
			err = client.Get(ctx, "/chain/head", nil, &head)
			if err != nil {
				return err
			}

			dlt := ask.Expiry - head.Height()
			rem := "<expired>"
			if dlt > 0 {
				rem = (time.Second * time.Duration(int64(dlt)*int64(constants.MainNetBlockDelaySecs))).String()
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%d\n", types.FIL(ask.Price), types.FIL(ask.VerifiedPrice), types.SizeStr(types.NewInt(uint64(ask.MinPieceSize))), types.SizeStr(types.NewInt(uint64(ask.MaxPieceSize))), ask.Expiry, rem, ask.SeqNo)
		}
		return w.Flush()
	},
}

var minerSetAskCmd = &cli.Command{
	Name:      "set",
	Usage:     "set miner ask",
	ArgsUsage: "[miner address]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "price",
			Usage:    "Set the price of the storage for unverified deals or retrievals deal (specified as FIL / GiB / Epoch) to `PRICE`.",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "verified-price",
			Usage: "Set the price of the storage ask for verified deals (specified as FIL / GiB / Epoch) to `PRICE`",
		},
		&cli.StringFlag{
			Name:        "min-piece-size",
			Usage:       "Set minimum piece size (w/bit-padding, in bytes) in storage ask to `SIZE`",
			DefaultText: "256B",
			Value:       "256B",
		},
		&cli.StringFlag{
			Name:        "max-piece-size",
			Usage:       "Set maximum piece size (w/bit-padding, in bytes) in storage ask to `SIZE`, eg. KiB, MiB, GiB, TiB, PiB",
			DefaultText: "miner sector size",
		},

		&cli.StringFlag{
			Name:  "unseal-price",
			Usage: "Set the price to unseal for retrieval",
		},
		&cli.StringFlag{
			Name:        "payment-interval",
			Usage:       "Set the payment interval (in bytes) for retrieval",
			DefaultText: "1MiB",
		},
		&cli.StringFlag{
			Name:        "payment-interval-increase",
			Usage:       "Set the payment interval increase (in bytes) for retrieval",
			DefaultText: "1MiB",
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		client, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.Args().Len() != 1 {
			return errors.New("must specify miner address")
		}
		mAddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return fmt.Errorf("para `miner` is invalid: %w", err)
		}

		if cctx.IsSet("retrieval") {
			ask := retrievalmarket.Ask{}
			err := client.Get(ctx, "/miner/ask/retrieval", mAddr, &ask)
			if err != nil {
				fmt.Println("error getting retrieval ask: ", err)
			}

			if cctx.IsSet("price") {
				v, err := types.ParseFIL(cctx.String("price"))
				if err != nil {
					return err
				}
				ask.PricePerByte = types.BigDiv(types.BigInt(v), types.NewInt(1<<30))
			}

			if cctx.IsSet("unseal-price") {
				v, err := types.ParseFIL(cctx.String("unseal-price"))
				if err != nil {
					return err
				}
				ask.UnsealPrice = abi.TokenAmount(v)
			}

			if cctx.IsSet("payment-interval") {
				v, err := units.RAMInBytes(cctx.String("payment-interval"))
				if err != nil {
					return err
				}
				ask.PaymentInterval = uint64(v)
			}

			if cctx.IsSet("payment-interval-increase") {
				v, err := units.RAMInBytes(cctx.String("payment-interval-increase"))
				if err != nil {
					return err
				}
				ask.PaymentIntervalIncrease = uint64(v)
			}

			req := service.MinerSetRetrievalAskReq{
				Ask:   ask,
				Miner: mAddr,
			}

			return client.Post(ctx, "/miner/ask/retrieval", &req, nil)
		}

		ask := storagemarket.StorageAsk{}
		err = client.Get(ctx, "/miner/ask/storage", mAddr, &ask)
		if err != nil {
			fmt.Println("error getting storage ask: ", err)
		}

		pri := ask.Price
		if cctx.IsSet("price") {
			v, err := types.ParseFIL(cctx.String("price"))
			if err != nil {
				return err
			}
			pri = types.BigInt(v)
		}

		vpri := ask.VerifiedPrice
		if cctx.IsSet("verified-price") {
			v, err := types.ParseFIL(cctx.String("verified-price"))
			if err != nil {
				return err
			}
			vpri = types.BigInt(v)
		}
		if vpri.Int == nil {
			vpri = pri
		}

		v, err := time.ParseDuration("720h0m0s")
		if err != nil {
			return err
		}
		dur := abi.ChainEpoch(v.Seconds() / float64(constants.MainNetBlockDelaySecs))

		min := ask.MinPieceSize
		if cctx.IsSet("min-piece-size") {
			v, err := units.RAMInBytes(cctx.String("min-piece-size"))
			if err != nil {
				return err
			}
			if v < 256 {
				return fmt.Errorf("min piece size must be at least 256 bytes")
			}
			min = abi.PaddedPieceSize(v)
		}

		max := ask.MaxPieceSize
		if cctx.IsSet("max-piece-size") {
			v, err := units.RAMInBytes(cctx.String("max-piece-size"))
			if err != nil {
				return err
			}
			max = abi.PaddedPieceSize(v)
		}

		req := service.MinerSetAskReq{
			Miner:         mAddr,
			Price:         pri,
			VerifiedPrice: vpri,
			MinPieceSize:  min,
			MaxPieceSize:  max,
			Duration:      dur,
		}

		return client.Post(cctx.Context, "/miner/ask/storage", req, nil)
	},
}
