package service

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

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
	"github.com/filecoin-project/venus/venus-shared/types"
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

func (s *ServiceImpl) MsgSend(ctx context.Context, params *MsgSendReq) (string, error) {

	decParams, err := params.Decode(s.Node)
	if err != nil {
		return "", err
	}

	msg := &types.Message{
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

func (s *ServiceImpl) MsgWait(ctx context.Context, msgId string) (*msgTypes.Message, error) {
	msg, err := s.Messager.WaitMessage(ctx, msgId, constants.DefaultConfidence)
	if err != nil {
		log.Errorf("wait message(%s) failed: %s", msgId, err)
		return nil, err
	}
	return msg, nil
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
			actor, err := s.Node.StateLookupID(ctx, params.From, types.EmptyTSK)
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
		msg := &types.Message{
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
	info, err := s.Node.StateMinerInfo(ctx, p.Miner, types.EmptyTSK)
	if err != nil {
		return fmt.Errorf("get miner sector size failed: %s", err)
	}

	smax := abi.PaddedPieceSize(info.SectorSize)

	if p.MaxPieceSize == 0 {
		p.MaxPieceSize = smax
	}

	if p.MaxPieceSize > smax {
		return fmt.Errorf("max piece size (w/bit-padding) %s cannot exceed miner sector size %s", types.SizeStr(types.NewInt(uint64(p.MaxPieceSize))), types.SizeStr(types.NewInt(uint64(smax))))
	}
	return s.Market.MarketSetAsk(ctx, p.Miner, p.Price, p.VerifiedPrice, p.Duration, p.MinPieceSize, p.MaxPieceSize)
}

func (s *ServiceImpl) MinerSetRetrievalAsk(ctx context.Context, p *MinerSetRetrievalAskReq) error {
	return s.Market.MarketSetRetrievalAsk(ctx, p.Miner, &p.Ask)
}

func (s *ServiceImpl) MinerInfo(ctx context.Context, mAddr address.Address) (*MinerInfoResp, error) {
	mi, err := s.Node.StateMinerInfo(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) info failed: %s", mAddr, err)
	}
	availBalance, err := s.Node.StateMinerAvailableBalance(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) available balance failed: %s", mAddr, err)
	}

	power, err := s.Node.StateMinerPower(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) power failed: %s", mAddr, err)
	}

	deadline, err := s.Node.StateMinerProvingDeadline(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) deadline failed: %s", mAddr, err)
	}

	return &MinerInfoResp{
		MinerInfo:    mi,
		MinerPower:   *power,
		AvailBalance: availBalance,
		Deadline:     *deadline,
	}, nil
}

func (s *ServiceImpl) MinerSetOwner(ctx context.Context, p *MinerSetOwnerReq) error {
	minerInfo, err := s.Node.StateMinerInfo(ctx, p.Miner, types.EmptyTSK)
	if err != nil {
		return fmt.Errorf("get miner(%s) info failed: %s", p.Miner, err)
	}

	newOwnerId, err := s.Node.StateLookupID(ctx, p.NewOwner, types.EmptyTSK)
	if err != nil {
		return fmt.Errorf("get new owner(%s) id failed: %s", p.NewOwner, err)
	}

	if minerInfo.Owner == newOwnerId {
		return fmt.Errorf("new owner(%s) is the same as old owner(%s)", p.NewOwner, minerInfo.Owner)
	}

	param, err := actors.SerializeParams(&newOwnerId)
	if err != nil {
		return fmt.Errorf("serialize params failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &types.Message{
		From:   minerInfo.Owner,
		To:     p.Miner,
		Method: builtin.MethodsMiner.ChangeOwnerAddress,
		Params: param,
		Value:  big.Zero(),
	}, nil)
	if err != nil {
		return fmt.Errorf("push message(%s) failed: %s", msg.ID, err)
	}

	return nil
}

func (s *ServiceImpl) MinerConfirmOwner(ctx context.Context, p *MinerSetOwnerReq) (oldOwner address.Address, err error) {
	minerInfo, err := s.Node.StateMinerInfo(ctx, p.Miner, types.EmptyTSK)
	if err != nil {
		return address.Undef, fmt.Errorf("get miner(%s) info failed: %s", p.Miner, err)
	}
	oldOwner = minerInfo.Owner

	newOwnerId, err := s.Node.StateLookupID(ctx, p.NewOwner, types.EmptyTSK)
	if err != nil {
		return address.Undef, fmt.Errorf("get new owner(%s) id failed: %s", p.NewOwner, err)
	}

	if minerInfo.Owner == newOwnerId {
		return address.Undef, fmt.Errorf("new owner(%s) is the same as old owner(%s)", p.NewOwner, minerInfo.Owner)
	}

	param, err := actors.SerializeParams(&newOwnerId)
	if err != nil {
		return address.Undef, fmt.Errorf("serialize params failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &types.Message{
		From:   p.NewOwner,
		To:     p.Miner,
		Method: builtin.MethodsMiner.ChangeOwnerAddress,
		Params: param,
		Value:  big.Zero(),
	}, nil)
	if err != nil {
		return address.Undef, fmt.Errorf("push message(%s) failed: %s", msg.ID, err)
	}

	return oldOwner, nil
}

func (s *ServiceImpl) MinerSetWorker(ctx context.Context, req *MinerSetWorkerReq) (WorkerChangeEpoch abi.ChainEpoch, err error) {
	minerInfo, err := s.Node.StateMinerInfo(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return 0, fmt.Errorf("get miner(%s) info failed: %s", req.Miner, err)
	}

	newWorkerId, err := s.Node.StateLookupID(ctx, req.NewWorker, types.EmptyTSK)
	if err != nil {
		return 0, fmt.Errorf("get new worker(%s) id failed: %s", req.NewWorker, err)
	}

	if minerInfo.Worker == newWorkerId {
		return 0, fmt.Errorf("new worker(%s) is the same as old worker(%s)", req.NewWorker, minerInfo.Worker)
	}

	if minerInfo.NewWorker == newWorkerId {
		return 0, fmt.Errorf("new worker(%s) has been proposed before, which will be effective after epoch(%d)", minerInfo.NewWorker, minerInfo.WorkerChangeEpoch)
	}

	param, err := actors.SerializeParams(&types.ChangeWorkerAddressParams{
		NewWorker:       newWorkerId,
		NewControlAddrs: minerInfo.ControlAddresses,
	})
	if err != nil {
		return 0, fmt.Errorf("serialize params failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &types.Message{
		From:   minerInfo.Owner,
		To:     req.Miner,
		Method: builtin.MethodsMiner.ChangeWorkerAddress,
		Params: param,
		Value:  big.Zero(),
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("push message(%s) failed: %s", msg.ID, err)
	}

	minerInfo, err = s.Node.StateMinerInfo(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return 0, fmt.Errorf("get miner(%s) info failed: %s", req.Miner, err)
	}

	return minerInfo.WorkerChangeEpoch, nil
}

func (s *ServiceImpl) MinerConfirmWorker(ctx context.Context, req *MinerSetWorkerReq) error {
	minerInfo, err := s.Node.StateMinerInfo(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return fmt.Errorf("get miner(%s) info failed: %s", req.Miner, err)
	}

	if minerInfo.NewWorker.Empty() {
		return fmt.Errorf("miner(%s) has no new worker", req.Miner)
	}

	if minerInfo.NewWorker != req.NewWorker {
		return fmt.Errorf("new worker(%s) is not the same as proposed worker(%s)", req.NewWorker, minerInfo.NewWorker)
	}

	head, err := s.Node.ChainHead(ctx)
	if err != nil {
		return fmt.Errorf("get chain head failed: %s", err)
	}

	if head.Height() < minerInfo.WorkerChangeEpoch {
		return fmt.Errorf("worker change epoch(%d) is not reached", minerInfo.WorkerChangeEpoch)
	}

	msg, err := s.PushMessageAndWait(ctx, &types.Message{
		From:   minerInfo.Owner,
		To:     req.Miner,
		Method: builtin.MethodsMiner.ConfirmUpdateWorkerKey,
		Value:  big.Zero(),
	}, nil)
	if err != nil {
		return fmt.Errorf("push message(%s) failed: %s", msg.ID, err)
	}

	return nil
}

func (s *ServiceImpl) MinerSetControllers(ctx context.Context, req *MinerSetControllersReq) (oldController []address.Address, err error) {
	minerInfo, err := s.Node.StateMinerInfo(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) info failed: %s", req.Miner, err)
	}
	oldController = minerInfo.ControlAddresses

	newControllers := make([]address.Address, 0, len(req.NewControllers))
	for _, c := range req.NewControllers {
		id, err := s.Node.StateLookupID(ctx, c, types.EmptyTSK)
		if err != nil {
			return nil, fmt.Errorf("get controller(%s) id failed: %s", c, err)
		}
		newControllers = append(newControllers, id)
	}

	if reflect.DeepEqual(minerInfo.ControlAddresses, newControllers) {
		return nil, fmt.Errorf("new controllers(%s) is the same as old controllers(%s)", req.NewControllers, minerInfo.ControlAddresses)
	}

	rawParam := &types.ChangeWorkerAddressParams{
		NewWorker:       minerInfo.Worker,
		NewControlAddrs: newControllers,
	}

	param, err := actors.SerializeParams(rawParam)
	if err != nil {
		return nil, fmt.Errorf("serialize params failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &types.Message{
		From:   minerInfo.Owner,
		To:     req.Miner,
		Method: builtin.MethodsMiner.ChangeWorkerAddress,
		Params: param,
		Value:  big.Zero(),
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("push message(%s) failed: %s", msg.ID, err)
	}

	return oldController, nil
}

func (s *ServiceImpl) MinerSetBeneficiary(ctx context.Context, req *MinerSetBeneficiaryReq) (*types.PendingBeneficiaryChange, error) {
	minerInfo, err := s.Node.StateMinerInfo(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) info failed: %s", req.Miner, err)
	}

	newBeneficiary, err := s.Node.StateLookupID(ctx, req.NewBeneficiary, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get beneficiary(%s) id failed: %s", req.NewBeneficiary, err)
	}
	req.NewBeneficiary = newBeneficiary

	if minerInfo.Beneficiary == newBeneficiary && newBeneficiary == minerInfo.Owner {
		return nil, fmt.Errorf("beneficiary %s already set to owner address", newBeneficiary)
	}

	param, err := actors.SerializeParams(&req.ChangeBeneficiaryParams)
	if err != nil {
		return nil, fmt.Errorf("serialize params failed: %s", err)
	}

	// owner proposal
	msg, err := s.PushMessageAndWait(ctx, &types.Message{
		From:   minerInfo.Owner,
		To:     req.Miner,
		Method: builtin.MethodsMiner.ChangeBeneficiary,
		Params: param,
		Value:  big.Zero(),
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("push message(%s) failed: %s", msg.ID, err)
	}

	minerInfo, err = s.Node.StateMinerInfo(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) info failed: %s", req.Miner, err)
	}

	if minerInfo.PendingBeneficiaryTerm == nil {
		return nil, fmt.Errorf("owner proposal beneficial change failed")
	}

	return minerInfo.PendingBeneficiaryTerm, nil
}

func (s *ServiceImpl) MinerConfirmBeneficiary(ctx context.Context, req *MinerConfirmBeneficiaryReq) (confirmor address.Address, err error) {
	minerInfo, err := s.Node.StateMinerInfo(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return address.Undef, fmt.Errorf("get miner(%s) info failed: %s", req.Miner, err)
	}

	if minerInfo.PendingBeneficiaryTerm == nil {
		return address.Undef, fmt.Errorf("miner(%s) no pending beneficiary", req.Miner)
	}
	if minerInfo.PendingBeneficiaryTerm.NewBeneficiary != req.NewBeneficiary {
		return address.Undef, fmt.Errorf("new beneficiary(%s) is not the same as proposed beneficiary(%s)", req.NewBeneficiary, minerInfo.PendingBeneficiaryTerm.NewBeneficiary)
	}

	sender := minerInfo.Beneficiary
	if !req.ByNominee {
		if minerInfo.PendingBeneficiaryTerm.ApprovedByBeneficiary {
			return address.Undef, fmt.Errorf("proposal already approved by beneficiary(%s)", minerInfo.Beneficiary)
		}
	} else {
		if minerInfo.PendingBeneficiaryTerm.ApprovedByNominee {
			return address.Undef, fmt.Errorf("proposal already approved by nominee(%s)", minerInfo.PendingBeneficiaryTerm.NewBeneficiary)
		}
		sender = minerInfo.PendingBeneficiaryTerm.NewBeneficiary
	}

	param, err := actors.SerializeParams(&types.ChangeBeneficiaryParams{
		NewBeneficiary: minerInfo.PendingBeneficiaryTerm.NewBeneficiary,
		NewQuota:       minerInfo.PendingBeneficiaryTerm.NewQuota,
		NewExpiration:  minerInfo.PendingBeneficiaryTerm.NewExpiration,
	})
	if err != nil {
		return address.Undef, fmt.Errorf("serialize params failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &types.Message{
		From:   sender,
		To:     req.Miner,
		Method: builtin.MethodsMiner.ChangeBeneficiary,
		Params: param,
		Value:  big.Zero(),
	}, nil)
	if err != nil {
		return address.Undef, fmt.Errorf("push message(%s) failed: %s", msg.ID, err)
	}

	return sender, nil
}

func (s *ServiceImpl) MinerGetDeadlines(ctx context.Context, mAddr address.Address) (*dline.Info, error) {
	return s.Node.StateMinerProvingDeadline(ctx, mAddr, types.EmptyTSK)
}

func (s *ServiceImpl) MinerWithdrawBalance(ctx context.Context, req *MinerWithdrawBalanceReq) (abi.TokenAmount, error) {

	minerInfo, err := s.Node.StateMinerInfo(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return big.Zero(), fmt.Errorf("get miner(%s) info failed: %s", req.Miner, err)
	}

	available, err := s.Node.StateMinerAvailableBalance(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return big.Zero(), fmt.Errorf("get miner(%s) available balance failed: %s", req.Miner, err)
	}

	if available.LessThan(req.Amount) {
		return big.Zero(), fmt.Errorf("withdraw amount(%s) is greater than available balance(%s)", req.Amount, available)
	}

	if req.Amount.LessThanEqual(big.Zero()) {
		req.Amount = available
	}

	param, err := actors.SerializeParams(&types.MinerWithdrawBalanceParams{
		AmountRequested: req.Amount,
	})
	if err != nil {
		return big.Zero(), fmt.Errorf("serialize params failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &types.Message{
		From:   minerInfo.Beneficiary,
		To:     req.Miner,
		Method: builtin.MethodsMiner.WithdrawBalance,
		Params: param,
		Value:  big.Zero(),
	}, nil)

	if err != nil {
		return big.Zero(), fmt.Errorf("push message(%s) failed: %s", msg.ID, err)
	}

	return req.Amount, nil
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
	rawParams := &types.ExtendSectorExpirationParams{}

	sectors := map[lminer.SectorLocation][]abi.SectorNumber{}
	for _, num := range req.SectorNumbers {
		p, err := s.Node.StateSectorPartition(ctx, req.Miner, num, types.EmptyTSK)
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
		rawParams.Extensions = append(rawParams.Extensions, types.ExpirationExtension{
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

	mi, err := s.Node.StateMinerInfo(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return fmt.Errorf("get miner info failed: %s", err)
	}

	_, err = s.Messager.PushMessage(ctx, &types.Message{
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
		sector, err := s.Node.StateSectorGetInfo(ctx, req.Miner, num, types.EmptyTSK)
		if err != nil {
			return nil, fmt.Errorf("get sector(%s) info failed: %s", num, err)
		}

		p, err := s.Node.StateSectorPartition(ctx, req.Miner, num, types.EmptyTSK)
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

func (s *ServiceImpl) PushMessageAndWait(ctx context.Context, msg *types.Message, spec *msgTypes.SendSpec) (*msgTypes.Message, error) {
	id, err := s.Messager.PushMessage(ctx, msg, spec)
	if err != nil {
		return nil, err
	}
	log.Infof("push message(%s) success", id)

	ret, err := s.MsgWait(ctx, id)
	if err != nil {
		return nil, err
	}
	log.Infof("messager(%s) is chained", id)

	return ret, nil
}
