package main

import (
	"reflect"

	"github.com/filecoin-project/venus/venus-devtool/api-gen/proxy"
	util "github.com/filecoin-project/venus/venus-devtool/util"
	"github.com/ipfs-force-community/venus-tool/service"
)

func main() {

	err := proxy.GenProxyForAPI(util.APIMeta{
		Type: reflect.TypeOf((*service.IService)(nil)).Elem(),
		ParseOpt: util.InterfaceParseOption{
			ImportPath: "github.com/ipfs-force-community/venus-tool/service",
			IncludeAll: true,
		},
	})

	if err != nil {
		panic(err)
	}

}
