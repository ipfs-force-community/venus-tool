package dep

import (
	"context"
	"fmt"

	"github.com/ipfs-force-community/damocles/damocles-manager/core"
	"github.com/ipfs-force-community/damocles/damocles-manager/dep"
	"github.com/ipfs-force-community/damocles/damocles-manager/pkg/workercli"
	"github.com/ipfs-force-community/venus-tool/repo/config"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"go.uber.org/fx"
)

type IDamocles interface {
	core.SealerCliAPI
}

type Damocles struct {
	core.SealerCliAPIClient
}

func NewDamocles(ctx context.Context, lc fx.Lifecycle, cfg *config.Config) (*Damocles, error) {
	if cfg.DamoclesAPI.Addr == "" {
		log.Warnf("damocles: %s", ErrEmptyAddr)
		return nil, nil
	}

	// transform the api addr from a multiaddr to a tcp addr
	ma, err := ma.NewMultiaddr(cfg.DamoclesAPI.Addr)
	if err != nil {
		return nil, err
	}
	_, addr, err := manet.DialArgs(ma)
	if err != nil {
		return nil, err
	}

	damocles := dep.MaybeAPIClient(ctx, lc, dep.ListenAddress(addr))

	return &Damocles{
		SealerCliAPIClient: damocles.SealerCliAPIClient,
	}, nil
}

type WorkerThreadInfo = core.WorkerThreadInfo
type WorkerPingInfo = core.WorkerPingInfo

type ThreadInfo struct {
	*core.WorkerThreadInfo
	WorkerInfo *core.WorkerInfo
	LastPing   int64
}

type WorkerClient struct {
	*core.WorkerPingInfo
	*workercli.Client
}

func NewWorkerClient(ctx context.Context, info *core.WorkerPingInfo) (*WorkerClient, func(), error) {
	c, closer, err := workercli.Connect(context.Background(), fmt.Sprintf("http://%s/", info.Info.Dest))
	if err != nil {
		return nil, nil, err
	}
	return &WorkerClient{
		WorkerPingInfo: info,
		Client:         c,
	}, closer, nil
}
