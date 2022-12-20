package service

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	nodeV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/api/market"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("service")

type Service struct {
	Messager messager.IMessager
	Market   market.IMarket
	Node     nodeV1.FullNode
	Wallets  []address.Address
	Miners   []address.Address
}

func (s *Service) MsgSend(ctx context.Context, params *SendReq) (string, error) {

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

func (s *Service) MsgQuery(ctx context.Context, params *QueryMsgReq) ([]*MsgResp, error) {
	var msgs []*msgTypes.Message
	var err error
	if params.ID != "" {
		msg, err := s.Messager.GetMessageByUid(ctx, params.ID)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	} else if params.Nonce != 0 {
		from, err := s.GetDefaultWallet(ctx)
		if err != nil {
			log.Warnf("get default wallet failed: %s", err)
		}
		if len(params.From) != 0 {
			from = params.From[0]
		}
		if from == address.Undef {
			return nil, fmt.Errorf("no sender indicated")
		}
		msg, err := s.Messager.GetMessageByFromAndNonce(ctx, from, params.Nonce)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	} else if params.IsBlocked {
		if len(params.From) == 0 {
			params.From = s.Wallets
		}
		for _, from := range params.From {
			msgsT, err := s.Messager.ListBlockedMessage(ctx, from, params.BlockedTime)
			if err != nil {
				log.Errorf("list blocked messages failed: %v", err)
			} else {
				msgs = append(msgs, msgsT...)
			}
		}
	} else if params.IsFailed {
		msgs, err = s.Messager.ListFailedMessage(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		msgs, err = s.Messager.ListMessage(ctx, &params.MsgQueryParams)
		if err != nil {
			return nil, err
		}
	}

	var ret []*MsgResp
	for _, msg := range msgs {
		resp := &MsgResp{
			Message: *msg,
		}
		_, err := resp.getMethodName(s.Node)
		if err != nil {
			log.Warnf("get method name failed: %s", err)
		}

		ret = append(ret, resp)
	}

	return ret, nil
}

func (s *Service) MsgReplace(ctx context.Context, params *msgTypes.ReplacMessageParams) (cid.Cid, error) {
	cid, err := s.Messager.ReplaceMessage(ctx, params)
	return cid, err
}

func (s *Service) GetDefaultWallet(ctx context.Context) (address.Address, error) {
	if len(s.Wallets) == 0 {
		return address.Undef, fmt.Errorf("no wallet configured")
	}
	return s.Wallets[0], nil
}
