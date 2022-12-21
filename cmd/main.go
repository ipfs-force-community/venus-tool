package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	logging "github.com/ipfs/go-log/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	NodeApi "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	MarketApi "github.com/filecoin-project/venus/venus-shared/api/market"
	msgApi "github.com/filecoin-project/venus/venus-shared/api/messager"

	vtCli "github.com/ipfs-force-community/venus-tool/cmd/cli"
	"github.com/ipfs-force-community/venus-tool/repo"
	"github.com/ipfs-force-community/venus-tool/repo/config"
	"github.com/ipfs-force-community/venus-tool/route"
	"github.com/ipfs-force-community/venus-tool/utils"
	"github.com/ipfs-force-community/venus-tool/version"

	"github.com/ipfs-force-community/venus-common-utils/builder"
	_ "github.com/ipfs-force-community/venus-tool/service"
)

var log = logging.Logger("main")

var flagRepo = &cli.StringFlag{
	Name:    "repo",
	EnvVars: []string{"VENUS_TOOL_PATH"},
	Aliases: []string{"tool-repo", "r"},
	Value:   "~/.venustool",
	Usage:   "Specify miner repo path, env VENUS_MINER_PATH",
}

var flagListen = &cli.StringFlag{
	Name:  "listen",
	Usage: "Specify the listen address",
	Value: "127.0.0.1:12580",
}

var flagNodeAPI = &cli.StringFlag{
	Name:    "node-api",
	Aliases: []string{"node"},
	Usage:   "specify venus node token and api address",
}

var flagMsgAPI = &cli.StringFlag{
	Name:    "msg-api",
	Aliases: []string{"msg"},
	Usage:   "specify venus-messager token and api address",
}

var flagMarketAPI = &cli.StringFlag{
	Name:    "market-api",
	Aliases: []string{"market"},
	Usage:   "specify venus-market token and api address",
}

var flagComToken = &cli.StringFlag{
	Name:    "common-token",
	Aliases: []string{"token"},
	Usage:   "specify venus common token",
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
		flagNodeAPI,
		flagMsgAPI,
		flagMarketAPI,
		flagComToken,
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
		if cfg.MessagerAPI.Addr == "" {
			return errors.New("messager api url is empty")
		}
		if cfg.MarketAPI.Addr == "" {
			return errors.New("market api url is empty")
		}
		if cfg.NodeAPI.Addr == "" {
			return errors.New("node api url is empty")
		}

		msgClient, msgCloser, err := msgApi.DialIMessagerRPC(ctx, cfg.MessagerAPI.Addr, cfg.MessagerAPI.Token, nil)
		if err != nil {
			return err
		}
		defer msgCloser()

		marketClient, marketCloser, err := MarketApi.DialIMarketRPC(ctx, cfg.MarketAPI.Addr, cfg.MarketAPI.Token, nil)
		if err != nil {
			return err
		}
		defer marketCloser()

		NodeClient, nodeCloser, err := NodeApi.DialFullNodeRPC(ctx, cfg.NodeAPI.Addr, cfg.NodeAPI.Token, nil)
		if err != nil {
			return err
		}
		defer nodeCloser()

		server := &http.Server{
			Addr:         cfg.Server.ListenAddr,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		}
		fx.Supply(server)

		// compose
		stop, err := builder.New(
			ctx,
			builder.Override(new(*http.Server), server),
			builder.Override(new(msgApi.IMessager), msgClient),
			builder.Override(new(MarketApi.IMarket), marketClient),
			builder.Override(new(NodeApi.FullNode), NodeClient),
			builder.Override(builder.NextInvoke(), utils.SetupLogLevels),
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

	updateApi := func(apiStr string, apiCfg *config.APIInfo) {
		addr, token := utils.ParseAPI(apiStr)
		if addr != "" {
			apiCfg.Addr = addr
		}
		if token != "" {
			apiCfg.Token = token
		}
	}

	if ctx.IsSet(flagListen.Name) {
		cfg.Server.ListenAddr = ctx.String(flagListen.Name)
	}

	if ctx.IsSet(flagNodeAPI.Name) {
		updateApi(ctx.String(flagNodeAPI.Name), cfg.NodeAPI)
	}
	if cfg.NodeAPI.Token == "" && commonToken != "" {
		cfg.NodeAPI.Token = commonToken
	}

	if ctx.IsSet(flagMsgAPI.Name) {
		updateApi(ctx.String(flagMsgAPI.Name), cfg.MessagerAPI)
	}
	if cfg.MessagerAPI.Token == "" && commonToken != "" {
		cfg.MessagerAPI.Token = commonToken
	}

	if ctx.IsSet(flagMarketAPI.Name) {
		updateApi(ctx.String(flagMarketAPI.Name), cfg.MarketAPI)
	}
	if cfg.MarketAPI.Token == "" && commonToken != "" {
		cfg.MarketAPI.Token = commonToken
	}
}
