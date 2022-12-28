package route

import (
	"context"
	"net/http"

	"github.com/ipfs-force-community/venus-tool/service"
	"go.uber.org/fx"
)

func RegisterAndStart(lc fx.Lifecycle, s service.Service, srv *http.Server) {
	srv.Handler = registerRoute(&s)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return srv.ListenAndServe()
		},
	})

}
