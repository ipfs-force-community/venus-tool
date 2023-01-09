package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/go-state-types/dline"
	power2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/power"
	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/filecoin-project/venus/venus-shared/actors"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin/miner"
	lminer "github.com/filecoin-project/venus/venus-shared/actors/builtin/miner"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin/power"
	nodeV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/api/market"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	marketTypes "github.com/filecoin-project/venus/venus-shared/types/market"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("service")

type ServiceImpl struct {
	Messager messager.IMessager
	Market   market.IMarket
	Node     nodeV1.FullNode
	Wallets  []address.Address
	Miners   []address.Address
}

var _ IService = &ServiceImpl{}

func (s *ServiceImpl) ChainHead(ctx context.Context) (*venusTypes.TipSet, error) {
	return s.Node.ChainHead(ctx)
}

func (s *ServiceImpl) GetDefaultWallet(ctx context.Context) (address.Address, error) {
	if len(s.Wallets) == 0 {
		return address.Undef, fmt.Errorf("no wallet configured")
	}
	return s.Wallets[0], nil
}

func (s *ServiceImpl) MsgSend(ctx context.Context, params *MsgSendReq) (string, error) {

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
		_, err := resp.getMethodName(s.Node)
		if err != nil {
			log.Warnf("get method name failed: %s", err)
		}

		ret = append(ret, resp)
	}

	return ret, nil
}

func (s *ServiceImpl) MsgReplace(ctx context.Context, params *MsgReplaceReq) (cid.Cid, error) {
	cid, err := s.Messager.ReplaceMessage(ctx, params)
	return cid, err
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

func (s *ServiceImpl) MinerCreate(ctx context.Context, params *MinerCreateReq) (address.Address, error) {
	// if target msg no exist, send it
	has := false
	var err error
	if params.MsgId != "" {
		has, err = s.Messager.HasMessageByUid(ctx, params.MsgId)
		if err != nil {
			return address.Undef, err
		}
	} else {
		params.MsgId = uuid.New().String()
	}

	if !has {
		sealProof, err := miner.SealProofTypeFromSectorSize(params.SectorSize, constants.TestNetworkVersion)
		if err != nil {
			return address.Undef, err
		}

		params.SealProofType = sealProof

		if params.Owner == address.Undef {
			actor, err := s.Node.StateLookupID(ctx, params.From, venusTypes.EmptyTSK)
			if err != nil {
				return address.Undef, err
			}
			params.Owner = actor
		}

		if params.Worker == address.Undef {
			params.Worker = params.Owner
		}

		p, err := actors.SerializeParams(&params.CreateMinerParams)
		if err != nil {
			return address.Undef, err
		}
		msg := &venusTypes.Message{
			From:   params.From,
			To:     power.Address,
			Method: power.Methods.CreateMiner,
			Params: p,
			Value:  big.Zero(),
		}

		_, err = s.Messager.PushMessageWithId(ctx, params.MsgId, msg, &msgTypes.SendSpec{})
		if err != nil {
			return address.Undef, err
		}
	}

	ret, err := s.Messager.GetMessageByUid(ctx, params.MsgId)
	if err != nil {
		return address.Undef, err
	}

	switch ret.State {
	case msgTypes.OnChainMsg, msgTypes.ReplacedMsg:
		if ret.Receipt.ExitCode != 0 {
			log.Warnf("message exec failed: %s(%d)", ret.Receipt.ExitCode, ret.Receipt.ExitCode)
			return address.Undef, fmt.Errorf("message exec failed: %s(%d)", ret.Receipt.ExitCode, ret.Receipt.ExitCode)
		}

		var cRes power2.CreateMinerReturn
		err = cRes.UnmarshalCBOR(bytes.NewReader(ret.Receipt.Return))
		if err != nil {
			return address.Undef, err
		}

		return cRes.IDAddress, nil

	case msgTypes.NoWalletMsg:
		log.Warnf("no wallet available for the sender %s, please check", params.From)
		return address.Undef, fmt.Errorf("no wallet available for the sender %s, please check", params.From)

	case msgTypes.FailedMsg:
		log.Warnf("message failed: %s", ret.ErrorMsg)
		return address.Undef, fmt.Errorf("message failed: %s", ret.ErrorMsg)

	default:
		log.Infof("msg state: %s", msgTypes.MessageStateToString(ret.State))
		return address.Undef, fmt.Errorf("temp error: waiting msg (%s) with state(%s) to be on chain", ret.ID, msgTypes.MessageStateToString(ret.State))
	}
}

func (s *ServiceImpl) MinerGetStorageAsk(ctx context.Context, mAddr address.Address) (*storagemarket.StorageAsk, error) {
	sAsk, err := s.Market.MarketGetAsk(ctx, mAddr)
	if err != nil {
		return nil, err
	}
	return sAsk.Ask, nil
}

func (s *ServiceImpl) MinerGetRetrievalAsk(ctx context.Context, mAddr address.Address) (*retrievalmarket.Ask, error) {
	return s.Market.MarketGetRetrievalAsk(ctx, mAddr)
}

func (s *ServiceImpl) MinerSetStorageAsk(ctx context.Context, p *MinerSetAskReq) error {
	info, err := s.Node.StateMinerInfo(ctx, p.Miner, venusTypes.EmptyTSK)
	if err != nil {
		return fmt.Errorf("get miner sector size failed: %s", err)
	}

	smax := abi.PaddedPieceSize(info.SectorSize)

	if p.MaxPieceSize == 0 {
		p.MaxPieceSize = smax
	}

	if p.MaxPieceSize > smax {
		return fmt.Errorf("max piece size (w/bit-padding) %s cannot exceed miner sector size %s", venusTypes.SizeStr(venusTypes.NewInt(uint64(p.MaxPieceSize))), venusTypes.SizeStr(venusTypes.NewInt(uint64(smax))))
	}
	return s.Market.MarketSetAsk(ctx, p.Miner, p.Price, p.VerifiedPrice, p.Duration, p.MinPieceSize, p.MaxPieceSize)
}

func (s *ServiceImpl) MinerSetRetrievalAsk(ctx context.Context, p *MinerSetRetrievalAskReq) error {
	return s.Market.MarketSetRetrievalAsk(ctx, p.Miner, &p.Ask)
}

func (s *ServiceImpl) MinerGetDeadlines(ctx context.Context, mAddr address.Address) (*dline.Info, error) {
	return s.Node.StateMinerProvingDeadline(ctx, mAddr, venusTypes.EmptyTSK)
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

func (s *ServiceImpl) SectorExtend(ctx context.Context, req SectorExtendReq) error {
	var err error
	rawParams := &venusTypes.ExtendSectorExpirationParams{}

	sectors := map[lminer.SectorLocation][]abi.SectorNumber{}
	for _, num := range req.SectorNumbers {
		p, err := s.Node.StateSectorPartition(ctx, req.Miner, num, venusTypes.EmptyTSK)
		if err != nil {
			return fmt.Errorf("get sector partition failed: %s", err)
		}

		if p == nil {
			return fmt.Errorf("sector %d not found", num)
		}

		sectors[*p] = append(sectors[*p], num)
	}

	for p, numbers := range sectors {
		nums := make([]uint64, len(numbers))
		for i, n := range numbers {
			nums[i] = uint64(n)
		}
		rawParams.Extensions = append(rawParams.Extensions, venusTypes.ExpirationExtension{
			Deadline:      p.Deadline,
			Partition:     p.Partition,
			Sectors:       bitfield.NewFromSet(nums),
			NewExpiration: req.Expiration,
		})
	}

	params, err := actors.SerializeParams(rawParams)
	if err != nil {
		return err
	}

	mi, err := s.Node.StateMinerInfo(ctx, req.Miner, venusTypes.EmptyTSK)
	if err != nil {
		return fmt.Errorf("get miner info failed: %s", err)
	}

	_, err = s.Messager.PushMessage(ctx, &venusTypes.Message{
		From:   mi.Worker,
		To:     req.Miner,
		Method: builtin.MethodsMiner.ExtendSectorExpiration,
		Params: params,
		Value:  big.Zero(),
	}, &msgTypes.SendSpec{})
	if err != nil {
		return fmt.Errorf("push message failed: %s", err)
	}

	return nil
}

func (s *ServiceImpl) SectorGet(ctx context.Context, req SectorGetReq) ([]*SectorResp, error) {
	ret := make([]*SectorResp, 0)
	for _, num := range req.SectorNumbers {
		sector, err := s.Node.StateSectorGetInfo(ctx, req.Miner, num, venusTypes.EmptyTSK)
		if err != nil {
			return nil, fmt.Errorf("get sector(%s) info failed: %s", num, err)
		}

		p, err := s.Node.StateSectorPartition(ctx, req.Miner, num, venusTypes.EmptyTSK)
		if err != nil {
			return nil, fmt.Errorf("get sector(%s) partition failed: %s", num, err)
		}

		ret = append(ret, &SectorResp{
			SectorOnChainInfo: *sector,
			SectorLocation:    *p,
		})
	}

	return ret, nil
}
