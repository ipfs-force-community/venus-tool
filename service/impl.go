package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin"

	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/filecoin-project/venus/pkg/state"
	nodeV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	market "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/filecoin-project/venus/venus-shared/blockstore"
	"github.com/filecoin-project/venus/venus-shared/types"
	marketTypes "github.com/filecoin-project/venus/venus-shared/types/market"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
	walletTypes "github.com/filecoin-project/venus/venus-shared/types/wallet"
	mkRepo "github.com/ipfs-force-community/droplet/v2/models/repo"
	minerTypes "github.com/ipfs-force-community/sophon-miner/types"
	"github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	logging "github.com/ipfs/go-log"
	cbg "github.com/whyrusleeping/cbor-gen"

	"github.com/ipfs-force-community/venus-tool/dep"
	"github.com/ipfs-force-community/venus-tool/pkg/multisig"
	"github.com/ipfs-force-community/venus-tool/utils"
)

var log = logging.Logger("service")

type ServiceImpl struct {
	Messager messager.IMessager
	Market   market.IMarket
	Node     nodeV1.FullNode
	Wallet   dep.IWallet
	Auth     dep.IAuth
	Damocles *dep.Damocles
	Miner    dep.Miner

	Multisig multisig.IMultiSig
}

var _ IService = &ServiceImpl{}

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
	wallets, err := s.Wallet.WalletList(ctx)
	if err != nil {
		return address.Undef, err
	}

	if len(wallets) == 0 {
		return address.Undef, fmt.Errorf("no wallet configured")
	}
	return wallets[0], nil
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
		case EncBase64:
			de, err := base64.StdEncoding.DecodeString(req.Data)
			if err != nil {
				return nil, err
			}
			return de, nil
		default:
			return nil, fmt.Errorf("unknown encoding type: %s", req.EncType)
		}
	}

	log.Infof("msg send: from(%s), to(%s), value(%s), method(%d), params(%s)", req.From, req.To, req.Value, req.Method, req.Params)

	var decParams []byte
	if req.Params != nil {
		var err error
		decParams, err = dec(*req.Params, req.To, req.Method)
		if err != nil {
			return "", fmt.Errorf("decode params failed: %s", err)
		}
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
			wallets, err := s.Wallet.WalletList(ctx)
			if err != nil {
				return nil, err
			}
			params.From = wallets
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
	for idx := range msgs {
		msg := msgs[idx]
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

		// decode params
		if len(msg.Params) != 0 {
			paramsRV := methodMeta.Params
			if paramsRV.Kind() == reflect.Ptr {
				paramsRV = paramsRV.Elem()
			}
			params := reflect.New(paramsRV).Interface().(cbg.CBORUnmarshaler)
			if err := params.UnmarshalCBOR(bytes.NewReader(msg.Params)); err != nil {
				log.Warnf("unmarshal params(%s) failed: %s", msg.Params, err)
			}
			p, err := json.MarshalIndent(params, "", "  ")
			if err != nil {
				log.Warnf("marshal params(%s) failed: %s", msg.Params, err)
			}
			resp.ParamsInJson = p
		}

		ret = append(ret, resp)
	}

	return ret, nil
}

func (s *ServiceImpl) Msg(ctx context.Context, id MsgID) (*MsgResp, error) {
	msg, err := s.Messager.GetMessageByUid(ctx, id.ID)
	if err != nil {
		return nil, fmt.Errorf("fail to get message by uid(%s): %s", id.ID, err)
	}
	ret := &MsgResp{
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
	ret.MethodName = methodMeta.Name

	// decode params
	if len(msg.Params) != 0 {
		paramsRV := methodMeta.Params
		if paramsRV.Kind() == reflect.Ptr {
			paramsRV = paramsRV.Elem()
		}
		params := reflect.New(paramsRV).Interface().(cbg.CBORUnmarshaler)
		if err := params.UnmarshalCBOR(bytes.NewReader(msg.Params)); err != nil {
			log.Warnf("unmarshal params(%s) failed: %s", msg.Params, err)
		}
		p, err := json.MarshalIndent(params, "", "  ")
		if err != nil {
			log.Warnf("marshal params(%s) failed: %s", msg.Params, err)
		}
		ret.ParamsInJson = p
	}

	return ret, err
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

func (s *ServiceImpl) MsgMarkBad(ctx context.Context, req *MsgID) error {
	return s.Messager.MarkBadMessage(ctx, req.ID)
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

func (s *ServiceImpl) AddrInfo(ctx context.Context, addr Address) (*AddrsResp, error) {
	if addr.Address.Empty() {
		return nil, fmt.Errorf("param error: address is empty")
	}
	var ret AddrsResp
	actorInfo, err := s.Node.GetActor(ctx, addr.Address)
	if err != nil {
		return nil, err
	}
	ret.Actor = *actorInfo

	addrInfo, err := s.Messager.GetAddress(ctx, addr.Address)
	if err != nil && strings.Contains(err.Error(), "not found") {
		ret.Address = msgTypes.Address{}
	} else if err != nil {
		return nil, err
	} else {
		ret.Address = *addrInfo
	}

	return &ret, nil

}

func (s *ServiceImpl) AddrList(ctx context.Context) ([]*AddrsResp, error) {
	addrInfos, err := s.Messager.ListAddress(ctx)
	if err != nil {
		return nil, err
	}

	addrs, err := s.Wallet.WalletList(ctx)
	if err != nil {
		return nil, err
	}
	allInfos := make([]*msgTypes.Address, 0, len(addrs)+len(addrInfos))
	for _, addr := range addrs {
		// if addr is not in addrInfos, add it
		var exist bool
		for _, addrInfo := range addrInfos {
			if addrInfo.Addr == addr {
				exist = true
				break
			}
		}
		if !exist {
			addrInfo, err := s.Messager.GetAddress(ctx, addr)
			if err != nil {
				log.Warnf("get address(%s) info failed: %s", addr, err)
				addrInfo = &msgTypes.Address{
					ID:   types.NewUUID(),
					Addr: addr,
				}
			}
			allInfos = append(allInfos, addrInfo)
		}
	}

	allInfos = append(allInfos, addrInfos...)

	ret := make([]*AddrsResp, 0, len(allInfos))
	for _, addrInfo := range allInfos {
		actorInfo, err := s.Node.GetActor(ctx, addrInfo.Addr)
		if err != nil {
			log.Warnf("get address(%s) actor failed: %s", addrInfo.Addr, err)
		}
		ret = append(ret, &AddrsResp{
			Address: *addrInfo,
			Actor:   *actorInfo,
		})
	}

	return ret, err
}

func (s *ServiceImpl) WalletList(ctx context.Context) ([]address.Address, error) {
	return s.Wallet.WalletList(ctx)
}

func (s *ServiceImpl) StorageDealList(ctx context.Context, miner Address) ([]marketTypes.MinerDeal, error) {
	if miner.Address != address.Undef {
		deals, err := s.Market.MarketListIncompleteDeals(ctx, &marketTypes.StorageDealQueryParams{Miner: miner.Address,
			Page: marketTypes.Page{
				Offset: 0, Limit: 100,
			},
		})
		if err != nil {
			return nil, err
		}
		return deals, nil
	}

	ret := make([]marketTypes.MinerDeal, 0)
	miners, err := s.listMiner(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range miners {

		deals, err := s.Market.MarketListIncompleteDeals(ctx, &marketTypes.StorageDealQueryParams{Miner: m})
		if err != nil {
			return nil, err
		}
		ret = append(ret, deals...)
	}
	return ret, nil
}

func (s *ServiceImpl) StorageDeal(ctx context.Context, proposalCid Cid) (*marketTypes.MinerDeal, error) {
	id, err := cid.Parse(proposalCid.Cid)
	if err != nil {
		return nil, fmt.Errorf("parse cid failed: %w", err)
	}
	return s.Market.MarketGetDeal(ctx, id)
}

func (s *ServiceImpl) StorageDealUpdateState(ctx context.Context, req StorageDealUpdateStateReq) error {
	return s.Market.UpdateStorageDealStatus(ctx, req.ProposalCid, req.State, req.PieceStatus)
}

func (s *ServiceImpl) RetrievalDealList(ctx context.Context) ([]marketTypes.ProviderDealState, error) {
	return s.Market.MarketListRetrievalDeals(ctx, &marketTypes.RetrievalDealQueryParams{})
}

func (s *ServiceImpl) WalletSignRecordQuery(ctx context.Context, req *WalletSignRecordQueryReq) ([]WalletSignRecordResp, error) {
	records, err := s.Wallet.ListSignedRecord(ctx, (*types.QuerySignRecordParams)(req))
	if err != nil {
		return nil, err
	}
	ret := make([]WalletSignRecordResp, 0, len(records))
	for _, r := range records {
		detail, err := GetDetailInJsonRawMessage(&r)
		if err != nil {
			return nil, err
		}
		ret = append(ret, WalletSignRecordResp{
			SignRecord: r,
			Detail:     detail,
		})
	}

	return ret, nil
}

func (s *ServiceImpl) Search(ctx context.Context, req SearchReq) (*SearchResp, error) {
	key := req.Key
	if key == "" {
		return nil, fmt.Errorf("param error: key is empty")
	}

	dataType := Unknown

	// judge key type
	if addr, err := address.NewFromString(key); err == nil {
		// key is address
		// address can be a wallet address or miner address
		head, err := s.Node.ChainHead(ctx)
		if err != nil {
			return nil, err
		}
		view := state.NewViewer(cbor.NewCborStore(blockstore.NewAPIBlockstore(s.Node))).StateView(head.ParentState())

		actor, err := view.LoadActor(ctx, addr)
		if err != nil {
			return nil, err
		}

		if builtin.IsStorageMinerActor(actor.Code) {
			// miner address
			dataType = Miner
		} else if builtin.IsAccountActor(actor.Code) {
			// wallet address
			dataType = Wallet
		} else {
			return nil, fmt.Errorf("address(%s) is not a miner or wallet address", addr)
		}
	} else {
		if mycid, err := cid.Decode(key); err == nil {
			// key is cid
			// cid can be proposal cid or messager id
			// try market
			_, err := s.Market.MarketGetDeal(ctx, mycid)
			if err == nil {
				// market deal
				dataType = Deal
			} else if !strings.Contains(err.Error(), mkRepo.ErrNotFound.Error()) {
				return nil, fmt.Errorf("check deal by cid(%s) failed: %s", key, err)
			}
		}

		// try messager
		has, err := s.Messager.HasMessageByUid(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("search message by uid(%s) failed: %s", key, err)
		}
		if has {
			dataType = Message
		}
	}

	if dataType == Unknown {
		return nil, fmt.Errorf("unknown key type")
	}

	ret := &SearchResp{
		Type: dataType,
	}

	switch dataType {
	case Miner:
		addr, _ := address.NewFromString(key) //lint:ignore SA1019 ignore err
		minerInfo, err := s.MinerInfo(ctx, Address{Address: addr})
		if err != nil {
			return nil, err
		}
		b, err := json.Marshal(minerInfo)
		if err != nil {
			return nil, err
		}
		ret.Data = json.RawMessage(b)
	case Wallet:
		addr, _ := address.NewFromString(key) //lint:ignore SA1019 ignore err
		walletInfo, err := s.AddrInfo(ctx, Address{Address: addr})
		if err != nil {
			return nil, err
		}
		b, err := json.Marshal(walletInfo)
		if err != nil {
			return nil, err
		}
		ret.Data = json.RawMessage(b)
	case Deal:
		deal, err := s.Market.MarketGetDeal(ctx, cid.MustParse(key))
		if err != nil {
			return nil, err
		}
		b, err := json.Marshal(deal)
		if err != nil {
			return nil, err
		}
		ret.Data = json.RawMessage(b)
	case Message:
		msg, err := s.Messager.GetMessageByUid(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("get message by uid %s error: %s", key, err)
		}
		b, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		ret.Data = json.RawMessage(b)
	}

	return ret, nil
}

func (s *ServiceImpl) MinedBlockList(ctx context.Context, req MinedBlockListReq) (MinedBlockListResp, error) {
	ret, err := s.Miner.ListBlocks(ctx, &minerTypes.BlocksQueryParams{
		Miners: req.Miner,
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil && !strings.Contains(err.Error(), "not support") {
		return MinedBlockListResp{}, err
	}
	return ret, nil
}

func (s *ServiceImpl) listMiner(ctx context.Context) ([]address.Address, error) {
	userName, err := s.Auth.GetUserName(ctx)
	if err != nil {
		return nil, err
	}

	miners, err := s.Auth.ListMiners(ctx, userName)
	if err != nil {
		return nil, err
	}

	ret := make([]address.Address, 0, len(miners))
	for _, m := range miners {
		ret = append(ret, m.Miner)
	}
	return ret, nil
}

func GetDetailInJsonRawMessage(r *types.SignRecord) (json.RawMessage, error) {
	t, ok := walletTypes.SupportedMsgTypes[r.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported type %s", r.Type)
	}

	wrap := func(err error) error {
		return fmt.Errorf("get detail: %w", err)
	}

	if r.RawMsg == nil {
		return nil, wrap(fmt.Errorf("msg is nil"))
	}

	if r.Type == types.MTVerifyAddress || r.Type == types.MTUnknown {
		// encode into hex string
		output := struct {
			Hex string
		}{
			Hex: hex.EncodeToString(r.RawMsg),
		}

		return json.Marshal(output)
	}

	signObj := reflect.New(t.Type).Interface()
	if err := walletTypes.CborDecodeInto(r.RawMsg, signObj); err != nil {
		return nil, fmt.Errorf("decode msg:%w", err)
	}
	return json.Marshal(signObj)
}
