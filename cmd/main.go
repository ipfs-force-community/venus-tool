package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	logging "github.com/ipfs/go-log/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	nodeApi "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	marketApi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	msgApi "github.com/filecoin-project/venus/venus-shared/api/messager"
	walletApi "github.com/filecoin-project/venus/venus-shared/api/wallet"
	"github.com/filecoin-project/venus/venus-shared/types"

	vtCli "github.com/ipfs-force-community/venus-tool/cmd/cli"
	"github.com/ipfs-force-community/venus-tool/dep"
	"github.com/ipfs-force-community/venus-tool/repo"
	"github.com/ipfs-force-community/venus-tool/repo/config"
	"github.com/ipfs-force-community/venus-tool/route"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/ipfs-force-community/venus-tool/utils"
	"github.com/ipfs-force-community/venus-tool/version"

	"github.com/ipfs-force-community/venus-common-utils/builder"
)

var log = logging.Logger("main")

var flagRepo = &cli.StringFlag{
	Name:    "repo",
	EnvVars: []string{"VENUS_TOOL_PATH"},
	Aliases: []string{"r"},
	Value:   "~/.venus-tool",
	Usage:   "Specify miner repo path, env VENUS_TOOL_PATH",
}

var flagListen = &cli.StringFlag{
	Name:  "listen",
	Usage: "Specify the listen address",
	Value: "127.0.0.1:8090",
}

var flagNodeAPI = &cli.StringFlag{
	Name:    "node-api",
	Aliases: []string{"node"},
	Usage:   "specify venus node token and api address. ex: --node-api=token:addr , if token was ignored, will use common token",
}

var flagMsgAPI = &cli.StringFlag{
	Name:    "msg-api",
	Aliases: []string{"msg"},
	Usage:   "specify venus-messager token and api address. ex: --msg-api=token:addr , if token was ignored, will use common token",
}

var flagMarketAPI = &cli.StringFlag{
	Name:    "market-api",
	Aliases: []string{"market"},
	Usage:   "specify venus-market token and api address. ex: --market-api=token:addr , if token was ignored, will use common token",
}

var flagWalletAPI = &cli.StringFlag{
	Name:    "wallet-api",
	Aliases: []string{"wallet"},
	Usage:   "specify venus-wallet token and api address. ex: --wallet-api=token:addr , if token was ignored, will use common token",
}

var flagAuthAPI = &cli.StringFlag{
	Name:    "auth-api",
	Aliases: []string{"auth"},
	Usage:   "specify venus-auth token and api address. ex: --auth-api=token:addr , if token was ignored, will use common token",
}

var flagMinerAPI = &cli.StringFlag{
	Name:    "miner-api",
	Aliases: []string{"miner"},
	Usage:   "specify venus-miner token and api address. ex: --miner-api=token:addr , if token was ignored, will use common token",
}

var flagDamoclesAPI = &cli.StringFlag{
	Name:    "damocles-api",
	Aliases: []string{"damocles"},
	Usage:   "specify venus-damocles token and api address. ex: --damocles-api=token:addr , if token was ignored, will use common token",
}

var flagComToken = &cli.StringFlag{
	Name:    "common-token",
	Aliases: []string{"token"},
	Usage:   "specify venus common token",
}

var flagDashboard = &cli.StringFlag{
	Name:    "board",
	Usage:   "specify path to static asset for dashboard",
	Value:   "./dashboard/build",
	EnvVars: []string{"SSM_DASHBOARD_PATH"},
}

func main() {
	app := &cli.App{
		Name:                 "venus-tool",
		Usage:                "tool for venus user to manage data on chain service , deal service and power service.",
		Version:              version.Version,
		Suggest:              true,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			flagRepo,
			vtCli.FlagServer,
		},
		Commands: []*cli.Command{
			runCmd,
			vtCli.MsgCmd,
			vtCli.AddrCmd,
			vtCli.MinerCmd,
			vtCli.DealCmd,
			vtCli.SectorCmd,
			vtCli.ChainCmd,
			vtCli.MultiSigCmd,
			vtCli.WalletCmd,
		},
	}
	app.Setup()
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		return
	}
}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "run venus-tool daemon",
	Flags: []cli.Flag{
		flagListen,
		flagAuthAPI,
		flagNodeAPI,
		flagMsgAPI,
		flagMarketAPI,
		flagMinerAPI,
		flagWalletAPI,
		flagDamoclesAPI,
		flagComToken,
		flagDashboard,
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		// load repo
		repoPath, err := homedir.Expand(cctx.String(flagRepo.Name))
		if err != nil {
			return err
		}
		repo := repo.NewRepo(repoPath)
		cfg := config.DefaultConfig()
		if repo.Exists() {
			cfg, err = repo.GetConfig()
			if err != nil {
				return err
			}
			updateFlag(cfg, cctx)
		} else {
			updateFlag(cfg, cctx)
			err := repo.Init(cfg)
			if err != nil {
				return err
			}
		}

		// todo replace it with stub
		if cfg.GetMessagerAPI().Addr == "" {
			return errors.New("messager api url is empty")
		}
		if cfg.GetMarketAPI().Addr == "" {
			return errors.New("market api url is empty")
		}
		if cfg.GetNodeAPI().Addr == "" {
			return errors.New("node api url is empty")
		}
		if cfg.WalletAPI.Addr == "" {
			return errors.New("wallet api url is empty")
		}

		msgClient, msgCloser, err := msgApi.DialIMessagerRPC(ctx, cfg.GetMessagerAPI().Addr, cfg.GetMessagerAPI().Token, nil)
		if err != nil {
			return err
		}
		defer msgCloser()

		marketClient, marketCloser, err := marketApi.DialIMarketRPC(ctx, cfg.GetMarketAPI().Addr, cfg.GetMarketAPI().Token, nil)
		if err != nil {
			return err
		}
		defer marketCloser()

		nodeClient, nodeCloser, err := nodeApi.DialFullNodeRPC(ctx, cfg.GetNodeAPI().Addr, cfg.GetNodeAPI().Token, nil)
		if err != nil {
			return err
		}
		defer nodeCloser()

		walletClient, walletCloser, err := walletApi.DialIFullAPIRPC(ctx, cfg.WalletAPI.Addr, cfg.WalletAPI.Token, nil)
		if err != nil {
			return err
		}
		defer walletCloser()

		server := &http.Server{
			Addr: cfg.Server.ListenAddr}
		fx.Supply(server)

		networkName, err := nodeClient.StateNetworkName(ctx)
		if err != nil {
			return err
		}

		// compose
		stop, err := builder.New(
			ctx,
			builder.Override(new(*config.Config), cfg),
			builder.Override(new(*http.Server), server),
			builder.Override(new(msgApi.IMessager), msgClient),
			builder.Override(new(marketApi.IMarket), marketClient),
			builder.Override(new(nodeApi.FullNode), nodeClient),
			builder.Override(new(dep.IWallet), walletClient),
			builder.Override(new(dep.IAuth), dep.NewAuth),
			builder.Override(new(*dep.Damocles), dep.NewDamocles),
			builder.Override(new(dep.Miner), dep.NewMiner),

			builder.Override(new(context.Context), ctx),
			builder.Override(new(types.NetworkName), networkName),
			builder.Override(new(*service.ServiceImpl), service.NewService),
			builder.Override(builder.NextInvoke(), utils.SetupLogLevels),
			builder.Override(builder.NextInvoke(), utils.LoadBuiltinActors),
			builder.Override(builder.NextInvoke(), route.RegisterAndStart),
		)
		if err != nil {
			return err
		}
		defer func() {
			log.Warn("received shutdown")

			log.Warn("Shutting down...")
			if err := stop(ctx); err != nil {
				log.Errorf("graceful shutting down failed: %s", err)
			}
			log.Info("Graceful shutdown successful")
		}()

		<-ctx.Done()
		return nil
	},
}

func updateFlag(cfg *config.Config, ctx *cli.Context) {

	commonToken := ctx.String(flagComToken.Name)

	boardPath := ctx.String(flagDashboard.Name)
	cfg.Server.BoardPath = boardPath
	// todo: parse relative path to absolute path

	updateApi := func(apiStr string, apiCfg *config.APIInfo) {
		if apiCfg == nil {
			apiCfg = &config.APIInfo{}
		}
		addr, token := utils.ParseAPI(apiStr)
		if addr != "" {
			apiCfg.Addr = addr
		}
		if token != "" {
			apiCfg.Token = token
		} else if commonToken != "" {
			apiCfg.Token = commonToken
		}
	}

	if ctx.IsSet(flagListen.Name) {
		cfg.Server.ListenAddr = ctx.String(flagListen.Name)
	}
	if ctx.IsSet(flagNodeAPI.Name) {
		updateApi(ctx.String(flagNodeAPI.Name), cfg.NodeAPI)
	}
	if ctx.IsSet(flagMsgAPI.Name) {
		updateApi(ctx.String(flagMsgAPI.Name), cfg.MessagerAPI)
	}
	if ctx.IsSet(flagMarketAPI.Name) {
		updateApi(ctx.String(flagMarketAPI.Name), cfg.MarketAPI)
	}
	if ctx.IsSet(flagWalletAPI.Name) {
		updateApi(ctx.String(flagWalletAPI.Name), &cfg.WalletAPI)
	}
	if ctx.IsSet(flagAuthAPI.Name) {
		updateApi(ctx.String(flagAuthAPI.Name), cfg.AuthAPI)
	}
	if ctx.IsSet(flagDamoclesAPI.Name) {
		updateApi(ctx.String(flagDamoclesAPI.Name), &cfg.DamoclesAPI)
	}
	if ctx.IsSet(flagMinerAPI.Name) {
		updateApi(ctx.String(flagMinerAPI.Name), cfg.MinerAPI)
	}
}
