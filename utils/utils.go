package utils

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/utils"
	"github.com/ipfs/go-cid"
)

type networkNameKeeper struct {
	NetworkName types.NetworkName
}

func (nk *networkNameKeeper) StateNetworkName(ctx context.Context) (types.NetworkName, error) {
	return nk.NetworkName, nil
}

func LoadBuiltinActors(ctx context.Context, networkName types.NetworkName) error {
	nk := &networkNameKeeper{NetworkName: networkName}
	return utils.LoadBuiltinActors(ctx, nk)
}

func GetMethodMeta(actorCode cid.Cid, method abi.MethodNum) (utils.MethodMeta, error) {
	methodMeta, found := utils.MethodsMap[actorCode][method]
	if !found {
		return utils.MethodMeta{}, fmt.Errorf("method %d not found on actor %s", method, actorCode)
	}
	return methodMeta, nil
}
