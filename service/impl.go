package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"

	"github.com/filecoin-project/venus/pkg/constants"
	nodeV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/api/market"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/filecoin-project/venus/venus-shared/types"
	marketTypes "github.com/filecoin-project/venus/venus-shared/types/market"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log"
	cbg "github.com/whyrusleeping/cbor-gen"

	"github.com/ipfs-force-community/venus-tool/pkg/multisig"
	"github.com/ipfs-force-community/venus-tool/utils"
)

var log = logging.Logger("service")

type ServiceImpl struct {
	Messager messager.IMessager
	Market   market.IMarket
	Node     nodeV1.FullNode
	Wallet   IWallet

	Multisig multisig.IMultiSig

	Wallets []address.Address
	Miners  []address.Address
}

var _ IService = &ServiceImpl{}

func (s *ServiceImpl) ChainGetHead(ctx context.Context) (*types.TipSet, error) {
	return s.Node.ChainHead(ctx)
}

func (s *ServiceImpl) ChainGetActor(ctx context.Context, addr address.Address) (*types.Actor, error) {
	return s.Node.StateGetActor(ctx, addr, types.EmptyTSK)
}

func (s *ServiceImpl) ChainGetNetworkName(ctx context.Context) (types.NetworkName, error) {
	return s.Node.StateNetworkName(ctx)
}

func (s *ServiceImpl) GetDefaultWallet(ctx context.Context) (address.Address, error) {
	if len(s.Wallets) == 0 {
		return address.Undef, fmt.Errorf("no wallet configured")
	}
	return s.Wallets[0], nil
}

func (s *ServiceImpl) MsgSend(ctx context.Context, req *MsgSendReq) (string, error) {
	dec := func(req EncodedParams, to address.Address, method abi.MethodNum) ([]byte, error) {
		switch req.EncType {
		case EncJson:
			act, err := s.Node.GetActor(ctx, to)
			if err != nil {
				return nil, err
			}
			return req.DecodeJSON(act.Code, method)
		case EncHex:
			return req.DecodeHex()
		case EncNull:
			return req.Data, nil
		default:
			return nil, fmt.Errorf("unknown encoding type: %s", req.EncType)
		}
	}

	decParams, err := dec(req.Params, req.To, req.Method)
	if err != nil {
		return "", err
	}

	msg := &types.Message{
		From:  req.From,
		To:    req.To,
		Value: req.Value,

		Method: req.Method,
		Params: decParams,
	}

	return s.Messager.PushMessage(ctx, msg, &req.SendSpec)
}

func (s *ServiceImpl) MsgQuery(ctx context.Context, params *MsgQueryReq) ([]*MsgResp, error) {
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

		act, err := s.Node.GetActor(ctx, msg.To)
		if err != nil {
			log.Warnf("get actor failed: %s", err)
		}
		methodMeta, err := utils.GetMethodMeta(act.Code, msg.Method)
		if err != nil {
			log.Warnf("get method meta failed: %s", err)
		}
		resp.MethodName = methodMeta.Name

		ret = append(ret, resp)
	}

	return ret, nil
}

func (s *ServiceImpl) MsgReplace(ctx context.Context, params *MsgReplaceReq) (cid.Cid, error) {
	cid, err := s.Messager.ReplaceMessage(ctx, params)
	return cid, err
}

func (s *ServiceImpl) MsgWait(ctx context.Context, msgId string) (*msgTypes.Message, error) {
	msg, err := s.Messager.WaitMessage(ctx, msgId, constants.DefaultConfidence)
	if err != nil {
		log.Errorf("wait message(%s) failed: %s", msgId, err)
		return nil, err
	}
	return msg, nil
}

func (s *ServiceImpl) MsgDecodeParam2Json(ctx context.Context, req *MsgDecodeParamReq) ([]byte, error) {
	if len(req.Params) == 0 {
		return []byte{}, nil
	}

	var err error

	act, err := s.Node.GetActor(ctx, req.To)
	if err != nil {
		return nil, err
	}
	methodMeta, err := utils.GetMethodMeta(act.Code, req.Method)
	if err != nil {
		return nil, err
	}

	paramsRV := methodMeta.Params
	if paramsRV.Kind() == reflect.Ptr {
		paramsRV = paramsRV.Elem()
	}
	params := reflect.New(paramsRV).Interface().(cbg.CBORUnmarshaler)
	if err := params.UnmarshalCBOR(bytes.NewReader(req.Params)); err != nil {
		return nil, err
	}
	return json.Marshal(params)
}

func (s *ServiceImpl) MsgGetMethodName(ctx context.Context, req *MsgGetMethodNameReq) (string, error) {
	act, err := s.Node.GetActor(ctx, req.To)
	if err != nil {
		return "", err
	}

	methodMeta, err := utils.GetMethodMeta(act.Code, req.Method)
	if err != nil {
		return "", err
	}

	return methodMeta.Name, nil
}

func (s *ServiceImpl) AddrOperate(ctx context.Context, params *AddrsOperateReq) error {
	has, err := s.Messager.HasAddress(ctx, params.Address)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("address not exist")
	}
	switch params.Operate {
	case DeleteAddress:
		return s.Messager.DeleteAddress(ctx, params.Address)
	case ActiveAddress:
		return s.Messager.ActiveAddress(ctx, params.Address)
	case ForbiddenAddress:
		return s.Messager.ForbiddenAddress(ctx, params.Address)
	case SetAddress:
		if params.IsSetSpec {
			err := s.Messager.SetFeeParams(ctx, &params.AddressSpec)
			if err != nil {
				return err
			}
		}
		if params.SelectMsgNum != 0 {
			return s.Messager.SetSelectMsgNum(ctx, params.Address, params.SelectMsgNum)
		}
		return nil
	default:
		return fmt.Errorf("unknown operate type")
	}
}

func (s *ServiceImpl) AddrList(ctx context.Context) ([]*AddrsResp, error) {
	addrs, err := s.Messager.ListAddress(ctx)
	if err != nil {
		return nil, err
	}
	ret := make([]*AddrsResp, 0, len(addrs))
	for _, addr := range addrs {
		ret = append(ret, (*AddrsResp)(addr))
	}
	return ret, err
}

func (s *ServiceImpl) StorageDealList(ctx context.Context, miner address.Address) ([]marketTypes.MinerDeal, error) {
	if miner != address.Undef {

		deals, err := s.Market.MarketListIncompleteDeals(ctx, miner)
		if err != nil {
			return nil, err
		}
		return deals, nil
	}

	ret := make([]marketTypes.MinerDeal, 0)
	for _, m := range s.Miners {
		deals, err := s.Market.MarketListIncompleteDeals(ctx, m)
		if err != nil {
			return nil, err
		}
		ret = append(ret, deals...)
	}
	return ret, nil
}

func (s *ServiceImpl) StorageDealUpdateState(ctx context.Context, req StorageDealUpdateStateReq) error {
	return s.Market.UpdateStorageDealStatus(ctx, req.ProposalCid, req.State, req.PieceStatus)
}

func (s *ServiceImpl) RetrievalDealList(ctx context.Context) ([]marketTypes.ProviderDealState, error) {
	return s.Market.MarketListRetrievalDeals(ctx)
}

func (s *ServiceImpl) PushMessageAndWait(ctx context.Context, msg *types.Message, spec *msgTypes.SendSpec) (*msgTypes.Message, error) {
	id, err := s.Messager.PushMessage(ctx, msg, spec)
	if err != nil {
		return nil, err
	}
	log.Infof("push message(%s) success", id)

	ret, err := s.MsgWait(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("wait message(%s) failed: %s", id, err)
	}

	log.Infof("messager(%s) is chained, exit code (%s), gas used(%d), return(%s)", id, ret.Receipt.ExitCode, ret.Receipt.GasUsed, len(ret.Receipt.Return))

	if ret.Receipt.ExitCode.IsError() {
		return ret, fmt.Errorf("exec messager failed: msgid(%s) exitcode(%s) return(%s)", ret.ID, ret.Receipt.ExitCode, ret.Receipt.Return)
	}
	return ret, nil
}
