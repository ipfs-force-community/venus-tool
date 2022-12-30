package service

import (
	"context"

	nodeV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/api/market"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"

	"go.uber.org/fx"
)

type Service struct {
	fx.In
	Messager messager.IMessager
	Market   market.IMarket
	Node     nodeV1.FullNode
}

func (s *Service) Send(ctx context.Context, params *SendParams) (string, error) {

	decParams, err := params.Decode(s.Node)
	if err != nil {
		return "", err
	}

	msg := &venusTypes.Message{
		From:  params.From,
		To:    params.To,
		Value: params.Value,

		Method: params.Method,
		Params: decParams,
	}

	return s.Messager.PushMessage(ctx, msg, &params.SendSpec)
}
