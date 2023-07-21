package main

import (
	"reflect"
	"testing"

	"github.com/filecoin-project/venus/venus-devtool/api-gen/proxy"
	util "github.com/filecoin-project/venus/venus-devtool/util"
	"github.com/ipfs-force-community/venus-tool/dep"
	"github.com/stretchr/testify/assert"
)

// func main() {

// 	err := proxy.GenProxyForAPI(util.APIMeta{
// 		Type: reflect.TypeOf((*service.IService)(nil)).Elem(),
// 		ParseOpt: util.InterfaceParseOption{
// 			ImportPath: "github.com/ipfs-force-community/venus-tool/service",
// 			IncludeAll: true,
// 		},
// 	})

// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Print("generate service success\n")

// 	err = proxy.GenProxyForAPI(util.APIMeta{
// 		Type: reflect.TypeOf((*dep.IDamocles)(nil)).Elem(),
// 		ParseOpt: util.InterfaceParseOption{
// 			ImportPath: "github.com/ipfs-force-community/venus-tool/dep",
// 			IncludeAll: true,
// 		},
// 	})

// 	if err != nil {
// 		panic(err)
// 	}
// }

func TestGenIDamocles(t *testing.T) {
	err := proxy.GenProxyForAPI(util.APIMeta{
		Type: reflect.TypeOf((*dep.IDamocles)(nil)).Elem(),
		ParseOpt: util.InterfaceParseOption{
			ImportPath: "github.com/ipfs-force-community/venus-tool/dep",
			IncludeAll: true,
		},
	})

	assert.NoError(t, err)
}
