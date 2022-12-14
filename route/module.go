package route

import (
	"net/http"

	"github.com/ipfs-force-community/venus-tool/service"
)

func RegisterAndStart(s service.Service, srv *http.Server) error {
	srv.Handler = RegisterRoute(&s)
	return srv.ListenAndServe()
}
