package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/libp2p/go-libp2p/core/peer"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
)

var MinerCmd = &cli.Command{
	Name:  "miner",
	Usage: "manage miner",
	Subcommands: []*cli.Command{
		minerInfoCmd,
		minerCreate,
		minerAskCmd,
		minerDeadlineCmd,
		minerSetOwnerCmd,
		minerSetWorkerCmd,
		minerSetControllersCmd,
		minerSetBeneficiaryCmd,
		minerWithdrawToBeneficiaryCmd,
		minerWithdrawFromMarketCmd,
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
	ArgsUsage: "<miner address>",
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context

		api, err := getAPI(cctx)
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
			ask, err := api.MinerGetRetrievalAsk(ctx, mAddr)
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

			ask, err := api.MinerGetStorageAsk(ctx, mAddr)
			if err != nil {
				return err
			}

			fmt.Fprintf(w, "Price per GiB/Epoch\tVerified\tMin. Piece Size (padded)\tMax. Piece Size (padded)\tExpiry (Epoch)\tExpiry (Appx. Rem. Time)\tSeq. No.\n")

			head, err := api.ChainGetHead(ctx)
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
	ArgsUsage: "<miner address>",
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
		api, err := getAPI(cctx)
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
			ask, err := api.MinerGetRetrievalAsk(ctx, mAddr)
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

			req := &service.MinerSetRetrievalAskReq{
				Ask:   *ask,
				Miner: mAddr,
			}

			err = api.MinerSetRetrievalAsk(ctx, req)
			if err != nil {
				return err
			}

			fmt.Println("retrieval ask updated")
			return nil
		}

		ask, err := api.MinerGetStorageAsk(ctx, mAddr)
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

		err = api.MinerSetStorageAsk(ctx, &req)
		if err != nil {
			return err
		}

		fmt.Println("storage ask updated")
		return nil
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
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		api, err := getAPI(cctx)
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

		fmt.Println("Creating miner, this might take a while")
		miner, err := api.MinerCreate(ctx, params)
		if err != nil {
			return err
		}

		fmt.Println(miner)
		return nil
	},
}

var minerDeadlineCmd = &cli.Command{
	Name:        "deadline",
	Usage:       "query miner proving deadline info",
	ArgsUsage:   "<Miner Address>",
	Description: `Query miner proving deadline info.`,
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		api, err := getAPI(cctx)
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

		dlInfo, err := api.MinerGetDeadlines(ctx, mAddr)
		if err != nil {
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

var minerInfoCmd = &cli.Command{
	Name:        "info",
	Usage:       "query miner info",
	ArgsUsage:   "<Miner Address>",
	Description: `Query miner info.`,
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		api, err := getAPI(cctx)
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

		mi, err := api.MinerInfo(ctx, mAddr)
		if err != nil {
			return err
		}

		return printJSON(mi)
	},
}

var minerSetOwnerCmd = &cli.Command{
	Name:      "set-owner",
	Usage:     "set the owner address of a miner (this command should be invoked by old owner firstly, then new owner invoke with '--confirm' flag to confirm the change)",
	ArgsUsage: "<minerAddress> <newOwnerAddress>",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "confirm",
			Usage: "confirm to change by the new owner",
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() != 2 {
			return fmt.Errorf("must pass miner address and new owner address as first and second arguments")
		}

		mAddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		newOwner, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		req := &service.MinerSetOwnerReq{
			Miner:    mAddr,
			NewOwner: newOwner,
		}

		fmt.Println("This will take some time (maybe 10 epoch), to ensure message is chained...")

		if cctx.Bool("confirm") {
			oldOwner, err := api.MinerConfirmOwner(ctx, req)
			if err != nil {
				return err
			}
			fmt.Printf("Miner owner changed to %s from %s \n", newOwner, oldOwner)
		} else {
			err = api.MinerSetOwner(ctx, req)
			if err != nil {
				return err
			}
			fmt.Printf("Miner owner proposed , it should be confirm by new owner(%s), who shall invoke 'set-owner' command with with '--confirm' flag \n", newOwner)
		}

		return nil
	},
}

var minerSetWorkerCmd = &cli.Command{
	Name:      "set-worker",
	Usage:     "set the worker address of a miner",
	ArgsUsage: "<minerAddress> <newWorkerAddress>",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "confirm",
			Usage:   "confirm the new worker address",
			Aliases: []string{"c"},
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() != 2 {
			return fmt.Errorf("must pass miner address and new worker address as first and second arguments")
		}

		mAddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		newWorker, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		req := &service.MinerSetWorkerReq{
			Miner:     mAddr,
			NewWorker: newWorker,
		}

		fmt.Println("This will take some time (maybe 5 epoch), to ensure message is chained...")

		if cctx.Bool("confirm") {
			err := api.MinerConfirmWorker(ctx, req)
			if err != nil {
				return err
			}
			fmt.Printf("Worker address changed to %s \n", newWorker)
			return nil
		}

		effectEpoch, err := api.MinerSetWorker(ctx, req)
		if err != nil {
			return err
		}

		fmt.Printf("Worker address(%s) change successfully proposed.\n", newWorker)
		fmt.Printf("Call 'set-worker' with '--confirm' flag at or after height %d to complete.\n", effectEpoch)

		return nil
	},
}

var minerSetControllersCmd = &cli.Command{
	Name:      "set-controllers",
	Usage:     "set the controllers of a miner",
	ArgsUsage: "<minerAddress> <newControllerAddresses>...",
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() < 2 {
			return fmt.Errorf("must pass miner address and at least one new controller address")
		}

		mAddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		var newControllers []address.Address
		add, del := map[address.Address]struct{}{}, map[address.Address]struct{}{}
		for _, a := range cctx.Args().Slice()[1:] {
			addr, err := address.NewFromString(a)
			if err != nil {
				return err
			}
			if _, ok := add[addr]; ok {
				return fmt.Errorf("duplicate address %s", addr)
			}
			add[addr] = struct{}{}
			newControllers = append(newControllers, addr)
		}

		req := &service.MinerSetControllersReq{
			Miner:          mAddr,
			NewControllers: newControllers,
		}

		fmt.Println("This will take some time (maybe 10 epoch), to ensure message is chained...")

		old, err := api.MinerSetControllers(ctx, req)
		if err != nil {
			return err
		}

		for _, a := range old {
			if _, ok := add[a]; ok {
				delete(add, a)
			} else {
				del[a] = struct{}{}
			}
		}

		if len(del) > 0 {
			fmt.Println("The following controllers are removed:")
			for a := range del {
				fmt.Println(a)
			}
		}

		if len(add) > 0 {
			fmt.Println("The following controllers are added:")
			for a := range add {
				fmt.Println(a)
			}
		}
		return nil
	},
}

var minerSetBeneficiaryCmd = &cli.Command{
	Name:      "set-beneficiary",
	Usage:     "set the beneficiary address of a miner (the change should be proposed by owner, and confirmed by old beneficiary and nominee)",
	ArgsUsage: "<minerAddress> <newBeneficiaryAddress> <quota> <expiration>",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "confirm-by-beneficiary",
			Usage: "confirm the change by the old beneficiary",
		},
		&cli.BoolFlag{
			Name:  "confirm-by-nominee",
			Usage: "confirm the change by the nominee",
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.Bool("confirm-by-beneficiary") && cctx.Bool("confirm-by-nominee") {
			return fmt.Errorf("can't confirm by beneficiary and nominee at the same time")
		}

		if cctx.NArg() < 2 {
			return fmt.Errorf("must pass miner address and new beneficiary address as first and second arguments")
		}

		mAddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		newBeneficiary, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		fmt.Println("This will take some time (maybe 5 epoch), to ensure message is chained...")
		if cctx.Bool("confirm-by-beneficiary") || cctx.Bool("confirm-by-nominee") {
			req := &service.MinerConfirmBeneficiaryReq{
				Miner:          mAddr,
				NewBeneficiary: newBeneficiary,
				ByNominee:      cctx.Bool("confirm-by-nominee"),
			}

			confirmor, err := api.MinerConfirmBeneficiary(ctx, req)
			if err != nil {
				return err
			}

			fmt.Printf("Beneficiary address changed to %s has been confirm by %s \n", newBeneficiary, confirmor)

		} else {
			if cctx.NArg() != 4 {
				return fmt.Errorf("must pass miner address, new beneficiary address, quota and expiration as arguments")
			}

			quota, err := types.ParseFIL(cctx.Args().Get(2))
			if err != nil {
				return err
			}

			expiration, err := strconv.ParseInt(cctx.Args().Get(3), 10, 64)
			if err != nil {
				return err
			}

			req := &service.MinerSetBeneficiaryReq{
				Miner: mAddr,
				ChangeBeneficiaryParams: types.ChangeBeneficiaryParams{
					NewBeneficiary: newBeneficiary,
					NewQuota:       abi.TokenAmount(quota),
					NewExpiration:  abi.ChainEpoch(expiration),
				},
			}

			pendingChange, err := api.MinerSetBeneficiary(ctx, req)
			if err != nil {
				return err
			}
			fmt.Println("Beneficiary change proposed:")
			err = printJSON(pendingChange)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

var minerWithdrawToBeneficiaryCmd = &cli.Command{
	Name:      "withdraw-to-beneficiary",
	Usage:     "withdraw balance from miner to beneficiary",
	ArgsUsage: "<minerAddress> <amount>",
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() != 2 {
			return fmt.Errorf("must pass miner address and amount as arguments")
		}

		mAddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		amount, err := types.ParseFIL(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		req := &service.MinerWithdrawBalanceReq{
			Miner:  mAddr,
			Amount: abi.TokenAmount(amount),
		}

		fmt.Println("This will take some time (maybe 5 epoch), to ensure message is chained...")
		withdrawn, err := api.MinerWithdrawToBeneficiary(ctx, req)
		if err != nil {
			return err
		}

		fmt.Printf("Balance of %s has been withdrawn \n", types.FIL(withdrawn))

		return nil
	},
}

var minerWithdrawFromMarketCmd = &cli.Command{
	Name:      "withdraw-from-market",
	Usage:     "withdraw balance from market",
	ArgsUsage: "<minerAddress> <amount>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "to",
			Usage: "the address will withdraw fund to, it should be the owner or worker of miner",
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		api, err := getAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.NArg() != 2 {
			return fmt.Errorf("must pass miner address and amount as arguments")
		}

		mAddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		amount, err := types.ParseFIL(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		req := &service.MinerWithdrawBalanceReq{
			Miner:  mAddr,
			Amount: abi.TokenAmount(amount),
		}

		if cctx.IsSet("to") {
			req.To, err = address.NewFromString(cctx.String("to"))
			if err != nil {
				return err
			}
		}

		fmt.Println("This will take some time (maybe 5 epoch), to ensure message is chained...")
		withdrawn, err := api.MinerWithdrawFromMarket(ctx, req)
		if err != nil {
			return err
		}

		fmt.Printf("Balance of %s has been withdrawn \n", types.FIL(withdrawn))

		return nil
	},
}
