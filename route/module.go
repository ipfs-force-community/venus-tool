package route

import (
	"context"
	"net/http"

	"github.com/ipfs-force-community/venus-tool/repo/config"
	"github.com/ipfs-force-community/venus-tool/service"
	"go.uber.org/fx"
)

func RegisterAndStart(lc fx.Lifecycle, s *service.ServiceImpl, srv *http.Server, cfg *config.Config) {
	srv.Handler = registerRoute(s, cfg.Server.BoardPath)
	log.Infof("load board from: %s", cfg.Server.BoardPath)
	log.Infof("server listen on: %s", cfg.Server.ListenAddr)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return srv.ListenAndServe()
		},
	})

}
