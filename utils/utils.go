package utils

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/utils"
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

func GetMethodMeta(node v1.IActor, to address.Address, method abi.MethodNum) (utils.MethodMeta, error) {
	ctx := context.Background()
	act, err := node.StateGetActor(ctx, to, types.EmptyTSK)
	if err != nil {
		return utils.MethodMeta{}, err
	}

	methodMeta, found := utils.MethodsMap[act.Code][method]
	if !found {
		return utils.MethodMeta{}, fmt.Errorf("method %d not found on actor %s", method, act.Code)
	}
	return methodMeta, nil
}
