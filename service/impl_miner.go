package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	rlepluslazy "github.com/filecoin-project/go-bitfield/rle"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/go-state-types/builtin/v11/power"
	"github.com/filecoin-project/go-state-types/dline"
	lminer "github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	lpower "github.com/filecoin-project/lotus/chain/actors/builtin/power"
	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/filecoin-project/venus/venus-shared/actors"
	"github.com/filecoin-project/venus/venus-shared/actors/adt"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin/miner"
	"github.com/filecoin-project/venus/venus-shared/blockstore"
	"github.com/filecoin-project/venus/venus-shared/types"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
	cbor "github.com/ipfs/go-ipld-cbor"
)

func (s *ServiceImpl) MinerCreate(ctx context.Context, params *MinerCreateReq) (address.Address, error) {
	wdProof, err := lminer.WindowPoStProofTypeFromSectorSize(params.SectorSize, constants.TestNetworkVersion)
	if err != nil {
		return address.Undef, err
	}

	params.WindowPoStProofType = wdProof

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

	msg, err := s.PushMessageAndWait(ctx, &types.Message{
		From:   params.From,
		To:     lpower.Address,
		Method: lpower.Methods.CreateMiner,
		Params: p,
		Value:  big.Zero(),
	}, nil)
	if err != nil {
		return address.Undef, fmt.Errorf("push message(%s) failed: %s", msg.ID, err)
	}

	var cRes power.CreateMinerReturn
	err = cRes.UnmarshalCBOR(bytes.NewReader(msg.Receipt.Return))
	if err != nil {
		return address.Undef, err
	}
	return cRes.IDAddress, nil
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

func (s *ServiceImpl) MinerInfo(ctx context.Context, addr Address) (*MinerInfoResp, error) {
	mAddr := addr.Address

	// load miner state
	mact, err := s.Node.StateGetActor(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}

	store := adt.WrapStore(ctx, cbor.NewCborStore(blockstore.NewAPIBlockstore(s.Node)))
	mst, err := miner.Load(store, mact)
	if err != nil {
		return nil, fmt.Errorf("load miner state: %w", err)
	}

	lockFund, err := mst.LockedFunds()
	if err != nil {
		return nil, err
	}

	mi, err := s.Node.StateMinerInfo(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) info failed: %w", mAddr, err)
	}
	availBalance, err := s.Node.StateMinerAvailableBalance(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) available balance failed: %w", mAddr, err)
	}

	power, err := s.Node.StateMinerPower(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) power failed: %w", mAddr, err)
	}

	deadline, err := s.Node.StateMinerProvingDeadline(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) deadline failed: %w", mAddr, err)
	}

	marketBalance, err := s.Node.StateMarketBalance(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}

	return &MinerInfoResp{
		MinerInfo:     mi,
		MinerPower:    *power,
		AvailBalance:  availBalance,
		Deadline:      *deadline,
		LockFunds:     lockFund,
		MarketBalance: marketBalance,
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
		Method: builtin.MethodsMiner.ConfirmChangeWorkerAddressExported,
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

	if newBeneficiary == minerInfo.Owner {
		req.NewQuota = big.Zero()
		req.NewExpiration = 0

		if minerInfo.Beneficiary == newBeneficiary {
			return nil, fmt.Errorf("beneficiary %s already set to owner address", newBeneficiary)
		}
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

func (s *ServiceImpl) MinerWithdrawToBeneficiary(ctx context.Context, req *MinerWithdrawBalanceReq) (abi.TokenAmount, error) {

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

func (s *ServiceImpl) MinerWithdrawFromMarket(ctx context.Context, req *MinerWithdrawBalanceReq) (abi.TokenAmount, error) {
	minerInfo, err := s.Node.StateMinerInfo(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return big.Zero(), fmt.Errorf("get miner(%s) info failed: %s", req.Miner, err)
	}
	marketBalance, err := s.Node.StateMarketBalance(ctx, req.Miner, types.EmptyTSK)
	if err != nil {
		return big.Zero(), fmt.Errorf("get miner(%s) available balance failed: %s", req.Miner, err)
	}

	reserved, err := s.Market.MarketGetReserved(ctx, req.Miner)
	if err != nil {
		return big.Zero(), fmt.Errorf("get miner(%s) reserved balance failed: %s", req.Miner, err)
	}

	avail := big.Subtract(big.Subtract(marketBalance.Escrow, marketBalance.Locked), reserved)

	if avail.LessThanEqual(big.Zero()) {
		return big.Zero(), fmt.Errorf("no available balance to withdraw")
	}

	if avail.LessThan(req.Amount) {
		return big.Zero(), fmt.Errorf("withdraw amount(%s) is greater than available balance(%s)", req.Amount, avail)
	}

	if req.Amount.LessThanEqual(big.Zero()) {
		req.Amount = avail
	}

	if req.To == address.Undef {
		req.To = minerInfo.Owner
	} else {
		toId, err := s.Node.StateLookupID(ctx, req.To, types.EmptyTSK)
		if err != nil {
			return big.Zero(), fmt.Errorf("lookup to address(%s) failed: %s", req.To, err)
		}
		req.To = toId
		if toId != minerInfo.Owner && toId != minerInfo.Worker {
			return big.Zero(), fmt.Errorf("to address(%s) is not miner owner(%s) or worker(%s)", req.To, minerInfo.Owner, minerInfo.Worker)
		}
	}

	mCid, err := s.Market.MarketWithdraw(ctx, req.To, req.Miner, req.Amount)
	if err != nil {
		return big.Zero(), fmt.Errorf("withdraw from market failed: %s", err)
	}
	id := mCid.String()
	log.Infof("push message(%s) success", id)

	msg, err := s.MsgWait(ctx, id)
	if err != nil {
		return big.Zero(), fmt.Errorf("push message(%s) failed: %s", msg.ID, err)
	}
	log.Infof("messager(%s) is chained", id)

	return req.Amount, nil
}

func (s *ServiceImpl) SectorExtend(ctx context.Context, req SectorExtendReq) error {
	var err error
	rawParams := &types.ExtendSectorExpirationParams{}

	sectors := map[miner.SectorLocation][]abi.SectorNumber{}
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

	_, err = s.PushMessageAndWait(ctx, &types.Message{
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

func (s *ServiceImpl) SectorList(ctx context.Context, req SectorListReq) ([]*types.SectorOnChainInfo, error) {
	mAddr := req.Miner

	// load miner state
	mact, err := s.Node.StateGetActor(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}

	store := adt.WrapStore(ctx, cbor.NewCborStore(blockstore.NewAPIBlockstore(s.Node)))
	mst, err := miner.Load(store, mact)
	if err != nil {
		return nil, fmt.Errorf("load miner state: %w", err)
	}

	pageSize, pageIndex := req.PageSize, req.PageIndex
	if pageSize == 0 {
		pageSize = 20
	}
	start := pageSize * pageIndex

	allocated, err := s.Node.StateMinerAllocated(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner(%s) allocated sectors failed: %s", mAddr, err)
	}

	iterator, err := allocated.BitIterator()
	if err != nil {
		return nil, fmt.Errorf("get iterator to range allocated sectors of miner(%s) failed: %s", mAddr, err)
	}

	ret := make([]*types.SectorOnChainInfo, 0)
	sectorNums := make([]uint64, 0, pageSize)

	n, err := iterator.Nth(uint64(start))
	if errors.Is(err, rlepluslazy.ErrEndOfIterator) {
		return ret, nil
	} else if err != nil {
		return nil, fmt.Errorf("get %dth sector number allocated of miner(%s) failed: %s", start, mAddr, err)
	}
	sectorNums = append(sectorNums, n)

	for i := 1; i < pageSize; i++ {
		if iterator.HasNext() {
			n, err := iterator.Next()
			if err != nil {
				return nil, fmt.Errorf("get %dth sector number allocated of miner(%s) failed: %s", start+i, mAddr, err)
			}
			sectorNums = append(sectorNums, n)
		}
	}
	log.Infof("sector numbers of miner(%s): %v", mAddr, sectorNums)

	for _, num := range sectorNums {
		sector, err := mst.GetSector(abi.SectorNumber(num))
		if sector == nil {
			if err != nil {
				log.Warnf("get sector(%d) info failed: %s", num, err)
			} else {
				log.Warnf("sector(%s) not found")
			}
			ret = append(ret, &types.SectorOnChainInfo{
				SectorNumber: abi.SectorNumber(num),
			})
			continue
		}
		ret = append(ret, sector)
	}
	return ret, nil
}

func (s *ServiceImpl) SectorSum(ctx context.Context, miner Address) (uint64, error) {
	allocated, err := s.Node.StateMinerAllocated(ctx, miner.Address, types.EmptyTSK)
	if err != nil {
		return 0, fmt.Errorf("get miner(%s) allocated sectors failed: %s", miner, err)
	}

	return allocated.Count()
}

func (s *ServiceImpl) MinerList(ctx context.Context) ([]address.Address, error) {
	return s.listMiner(ctx)
}

func (s *ServiceImpl) MinerWinCount(ctx context.Context, req *MinerWinCountReq) (MinerWinCountResp, error) {
	if s.Miner == nil {
		return MinerWinCountResp{}, ErrEmptyMiner
	}
	// todo: cache the result
	return s.Miner.CountWinners(ctx, req.Miners, req.From, req.To)
}
