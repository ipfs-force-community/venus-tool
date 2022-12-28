package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/dline"
	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/google/uuid"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/libp2p/go-libp2p/core/peer"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
)

var MinerCmd = &cli.Command{
	Name:  "miner",
	Usage: "manage miner",
	Subcommands: []*cli.Command{
		minerCreate,
		minerAskCmd,
		minerDeadlineCmd,
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

var minerCreate = &cli.Command{
	Name: "create",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Required: true,
			Usage:    "Wallet address used for sending the `CreateMiner` message",
		},
		&cli.StringFlag{
			Name:  "owner",
			Usage: "Actor address used as the `Owner` field. Will use the AccountActor ID of `from` if not provided",
		},
		&cli.StringFlag{
			Name:  "worker",
			Usage: "Actor address used as the `Worker` field. Will use the AccountActor ID of `from` if not provided",
		},
		&cli.StringFlag{
			Name:     "sector-size",
			Required: true,
			Usage:    "Sector size of the miner, 512MiB, 32GiB, 64GiB, etc",
		},
		&cli.StringFlag{
			Name:  "peer",
			Usage: "P2P peer id of the miner",
		},
		&cli.StringSliceFlag{
			Name:  "multiaddr",
			Usage: "P2P peer address of the miner",
		},
		&cli.StringFlag{
			Name:  "exid",
			Usage: "extra identifier to avoid duplicate msg id for pushing into venus-messager",
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		client, err := getAPI(cctx)
		if err != nil {
			return err
		}

		params := &service.MinerCreateReq{}

		ssize, err := units.RAMInBytes(cctx.String("sector-size"))
		if err != nil {
			return fmt.Errorf("failed to parse sector size: %w", err)
		}

		params.SectorSize = abi.SectorSize(ssize)

		fromStr := cctx.String("from")
		from, err := address.NewFromString(fromStr)
		if err != nil {
			return fmt.Errorf("parse from addr %s: %w", fromStr, err)
		}

		params.From = from

		if cctx.IsSet("owner") {
			ownerStr := cctx.String("owner")
			owner, err := address.NewFromString(ownerStr)
			if err != nil {
				return fmt.Errorf("parse owner addr %s: %w", ownerStr, err)
			}
			params.Owner = owner
		}

		if s := cctx.String("worker"); s != "" {
			addr, err := address.NewFromString(s)
			if err != nil {
				return fmt.Errorf("parse worker addr %s: %w", s, err)
			}

			params.Worker = addr
		}

		if cctx.IsSet("peer") {
			s := cctx.String("peer")

			id, err := peer.Decode(s)
			if err != nil {
				return fmt.Errorf("parse peer id %s: %w", s, err)
			}

			params.Peer = abi.PeerID(id)
		}

		if cctx.IsSet("multiaddr") {
			for _, one := range cctx.StringSlice("multiaddr") {
				maddr, err := ma.NewMultiaddr(one)
				if err != nil {
					return fmt.Errorf("parse multiaddr %s: %w", one, err)
				}

				maddrNop2p, strip := ma.SplitFunc(maddr, func(c ma.Component) bool {
					return c.Protocol().Code == ma.P_P2P
				})

				if strip != nil {
					fmt.Println("Stripping peerid ", strip, " from ", maddr)
				}

				params.Multiaddrs = append(params.Multiaddrs, maddrNop2p.Bytes())
			}
		}

		if cctx.IsSet("exid") {
			params.MsgId = cctx.String("exid")
		} else {
			params.MsgId = uuid.New().String()
		}

		miner := address.Address{}
		err = client.Post(ctx, "/miner/create", params, &miner)

		for err != nil && strings.Contains(err.Error(), "temp error") {
			log.Debugf("on waiting: %s", err)
			time.Sleep(5 * time.Second)
			err = client.Post(ctx, "/miner/create", map[string]string{"MsgId": params.MsgId}, &miner)
		}

		fmt.Println(miner)

		return nil
	},
}

var minerDeadlineCmd = &cli.Command{
	Name:      "deadline",
	Usage:     "query miner proving deadline info",
	ArgsUsage: "[minerAddress]",
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		client, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() != 1 {
			return fmt.Errorf("must pass miner address as first and only argument")
		}

		mAddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		var dlInfo *dline.Info
		if err := client.Get(ctx, "/miner/deadline", mAddr, &dlInfo); err != nil {
			return err
		}

		fmt.Printf("Period Start:\t%s\n", dlInfo.PeriodStart)
		fmt.Printf("Index:\t\t%d\n", dlInfo.Index)
		fmt.Printf("Open:\t\t%s\n", dlInfo.Open)
		fmt.Printf("Close:\t\t%s\n", dlInfo.Close)
		fmt.Printf("Challenge:\t%s\n", dlInfo.Challenge)
		fmt.Printf("FaultCutoff:\t%s\n", dlInfo.FaultCutoff)

		return nil

	},
}
