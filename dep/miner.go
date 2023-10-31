package dep

import (
	"context"
	"fmt"

	"github.com/ipfs-force-community/sophon-miner/api"
	"github.com/ipfs-force-community/sophon-miner/api/client"
	"github.com/ipfs-force-community/venus-tool/repo/config"
	"go.uber.org/fx"
)

type Miner api.MinerAPI

func NewMiner(ctx context.Context, lc fx.Lifecycle, cfg *config.Config) (Miner, error) {
	if cfg.MinerAPI.Addr == "" {
		log.Warnf("miner: %s", ErrEmptyAddr)
		return nil, nil
	}

	entryPoint, err := cfg.MinerAPI.DialArgs("v0")
	if err != nil {
		return nil, err
	}

	header := cfg.MinerAPI.AuthHeader()
	if header == nil {
		return nil, fmt.Errorf("gen auth header fail")
	}
	fmt.Print(header)

	api, closer, err := client.NewMinerRPC(ctx, entryPoint, header)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			closer()
			return nil
		},
	})

	return api, nil

}
