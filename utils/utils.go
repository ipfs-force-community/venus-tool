package utils

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/utils"
)

func LoadBuiltinActors(ctx context.Context, node v1.FullNode) error {
	return utils.LoadBuiltinActors(ctx, node)
}

func GetMethodMeta(node v1.IActor, to address.Address, method abi.MethodNum) (utils.MethodMeta, error) {
	ctx := context.Background()
	act, err := node.StateGetActor(ctx, to, venusTypes.EmptyTSK)
	if err != nil {
		return utils.MethodMeta{}, err
	}

	methodMeta, found := utils.MethodsMap[act.Code][method]
	if !found {
		return utils.MethodMeta{}, fmt.Errorf("method %d not found on actor %s", method, act.Code)
	}
	return methodMeta, nil
}
