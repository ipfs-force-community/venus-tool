package route

import (
	"testing"

	"github.com/ipfs-force-community/venus-tool/service"
)

func TestRouteInfo(t *testing.T) {
	routeInfos := Parse(service.IServiceStruct{}.Internal)
	t.Logf("%+v", routeInfos)
}
