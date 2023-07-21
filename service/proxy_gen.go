// Code generated by github.com/filecoin-project/venus/venus-devtool/api-gen. DO NOT EDIT.
package service

import (
	"context"

	address "github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/dline"
	"github.com/ipfs-force-community/venus-tool/dep"
	cid "github.com/ipfs/go-cid"

	"github.com/filecoin-project/venus/venus-shared/types"
	marketTypes "github.com/filecoin-project/venus/venus-shared/types/market"
)

type IServiceStruct struct {
	Internal struct {
		AddrInfo                   func(ctx context.Context, addr Address) (*AddrsResp, error)                                         ` GET:"/addr/info/:Address"`
		AddrList                   func(ctx context.Context) ([]*AddrsResp, error)                                                     ` GET:"/addr/list"`
		AddrOperate                func(ctx context.Context, params *AddrsOperateReq) error                                            ` PUT:"/addr/operate"`
		ChainGetActor              func(ctx context.Context, addr address.Address) (*types.Actor, error)                               ` GET:"/chain/actor"`
		ChainGetHead               func(ctx context.Context) (*types.TipSet, error)                                                    ` GET:"/chain/head"`
		ChainGetNetworkName        func(ctx context.Context) (types.NetworkName, error)                                                ` GET:"/chain/networkname"`
		MinedBlockList             func(ctx context.Context, req MinedBlockListReq) (MinedBlockListResp, error)                        ` GET:"/minedblock/list"`
		MinerConfirmBeneficiary    func(ctx context.Context, req *MinerConfirmBeneficiaryReq) (confirmor address.Address, err error)   ` PUT:"/miner/confirmbeneficiary"`
		MinerConfirmOwner          func(ctx context.Context, p *MinerSetOwnerReq) (oldOwner address.Address, err error)                ` PUT:"/miner/confirmowner"`
		MinerConfirmWorker         func(ctx context.Context, req *MinerSetWorkerReq) error                                             ` PUT:"/miner/confirmworker"`
		MinerCreate                func(ctx context.Context, params *MinerCreateReq) (address.Address, error)                          ` POST:"/miner/create"`
		MinerGetDeadlines          func(ctx context.Context, mAddr address.Address) (*dline.Info, error)                               ` GET:"/miner/deadline"`
		MinerGetRetrievalAsk       func(ctx context.Context, mAddr address.Address) (*retrievalmarket.Ask, error)                      ` GET:"/miner/retrievalask"`
		MinerGetStorageAsk         func(ctx context.Context, mAddr address.Address) (*storagemarket.StorageAsk, error)                 ` GET:"/miner/storageask"`
		MinerInfo                  func(ctx context.Context, mAddr Address) (*MinerInfoResp, error)                                    ` GET:"/miner/info/:Address"`
		MinerList                  func(ctx context.Context) ([]address.Address, error)                                                ` GET:"/miner/list"`
		MinerSetBeneficiary        func(ctx context.Context, req *MinerSetBeneficiaryReq) (*types.PendingBeneficiaryChange, error)     ` PUT:"/miner/beneficiary"`
		MinerSetControllers        func(ctx context.Context, req *MinerSetControllersReq) (oldController []address.Address, err error) ` PUT:"/miner/controllers"`
		MinerSetOwner              func(ctx context.Context, p *MinerSetOwnerReq) error                                                ` PUT:"/miner/owner"`
		MinerSetRetrievalAsk       func(ctx context.Context, p *MinerSetRetrievalAskReq) error                                         ` PUT:"/miner/retrievalask"`
		MinerSetStorageAsk         func(ctx context.Context, p *MinerSetAskReq) error                                                  ` PUT:"/miner/storageask"`
		MinerSetWorker             func(ctx context.Context, req *MinerSetWorkerReq) (WorkerChangeEpoch abi.ChainEpoch, err error)     ` PUT:"/miner/worker"`
		MinerWinCount              func(ctx context.Context, req *MinerWinCountReq) (MinerWinCountResp, error)                         ` GET:"/miner/wincount"`
		MinerWithdrawFromMarket    func(ctx context.Context, req *MinerWithdrawBalanceReq) (abi.TokenAmount, error)                    ` PUT:"/miner/withdrawmarket"`
		MinerWithdrawToBeneficiary func(ctx context.Context, req *MinerWithdrawBalanceReq) (abi.TokenAmount, error)                    ` PUT:"/miner/withdrawbeneficiary"`
		Msg                        func(ctx context.Context, id MsgID) (*MsgResp, error)                                               ` GET:"/msg/:ID"`
		MsgDecodeParam2Json        func(ctx context.Context, req *MsgDecodeParamReq) ([]byte, error)                                   ` POST:"/msg/decodeparam"`
		MsgGetMethodName           func(ctx context.Context, req *MsgGetMethodNameReq) (string, error)                                 ` GET:"/msg/getmethodname"`
		MsgMarkBad                 func(ctx context.Context, req *MsgID) error                                                         ` POST:"/msg/markbad/:ID"`
		MsgQuery                   func(ctx context.Context, params *MsgQueryReq) ([]*MsgResp, error)                                  ` GET:"/msg/query"`
		MsgReplace                 func(ctx context.Context, params *MsgReplaceReq) (cid.Cid, error)                                   ` POST:"/msg/replace"`
		MsgSend                    func(ctx context.Context, params *MsgSendReq) (string, error)                                       ` POST:"/msg/send"`
		MsigAddSigner              func(ctx context.Context, req *MultisigChangeSignerReq) (*types.ProposeReturn, error)               ` POST:"/msig/signer/ass"`
		MsigApprove                func(ctx context.Context, req *MultisigApproveReq) (*types.ApproveReturn, error)                    ` POST:"/msig/approve"`
		MsigCancel                 func(ctx context.Context, req *MultisigCancelReq) error                                             ` POST:"/msig/cancel"`
		MsigCreate                 func(ctx context.Context, req *MultisigCreateReq) (address.Address, error)                          ` POST:"/msig/create"`
		MsigInfo                   func(ctx context.Context, msig address.Address) (*types.MsigInfo, error)                            ` GET:"/msig/info"`
		MsigListPropose            func(ctx context.Context, msig address.Address) ([]*types.MsigTransaction, error)                   ` GET:"/msig/proposes"`
		MsigPropose                func(ctx context.Context, req *MultisigProposeReq) (*types.ProposeReturn, error)                    ` POST:"/msig/propose"`
		MsigRemoveSigner           func(ctx context.Context, req *MultisigChangeSignerReq) (*types.ProposeReturn, error)               ` POST:"/msig/signer/remove"`
		MsigSwapSigner             func(ctx context.Context, req *MultisigSwapSignerReq) (*types.ProposeReturn, error)                 ` POST:"/msig/signer/swap"`
		RetrievalDealList          func(ctx context.Context) ([]marketTypes.ProviderDealState, error)                                  ` GET:"/deal/retrieval"`
		Search                     func(ctx context.Context, req SearchReq) (*SearchResp, error)                                       ` GET:"/search/:Key"`
		SectorExtend               func(ctx context.Context, req SectorExtendReq) error                                                ` PUT:"/sector/extend"`
		SectorGet                  func(ctx context.Context, req SectorGetReq) ([]*SectorResp, error)                                  ` GET:"/sector/get"`
		SectorList                 func(ctx context.Context, req SectorListReq) ([]*types.SectorOnChainInfo, error)                    ` GET:"/sector/list"`
		SectorSum                  func(ctx context.Context, miner Address) (uint64, error)                                            ` GET:"/sector/sum"`
		StorageDeal                func(ctx context.Context, proposalCid Cid) (*marketTypes.MinerDeal, error)                          ` GET:"/deal/storage/info/:Cid"`
		StorageDealList            func(ctx context.Context, miner Address) ([]marketTypes.MinerDeal, error)                           ` GET:"/deal/storage/:Address"`
		StorageDealUpdateState     func(ctx context.Context, req StorageDealUpdateStateReq) error                                      ` PUT:"/deal/storage/state"`
		ThreadList                 func(ctx context.Context) ([]*dep.ThreadInfo, error)                                                ` GET:"/thread/list"`
		ThreadStart                func(ctx context.Context, req *ThreadStartReq) error                                                ` PUT:"/thread/start"`
		ThreadStop                 func(ctx context.Context, req *ThreadStopReq) error                                                 ` PUT:"/thread/stop"`
		WalletList                 func(ctx context.Context) ([]address.Address, error)                                                ` GET:"/wallet/list"`
		WalletSignRecordQuery      func(ctx context.Context, req *WalletSignRecordQueryReq) ([]WalletSignRecordResp, error)            ` GET:"/wallet/signrecord"`
	}
}

func (s *IServiceStruct) AddrInfo(p0 context.Context, p1 Address) (*AddrsResp, error) {
	return s.Internal.AddrInfo(p0, p1)
}
func (s *IServiceStruct) AddrList(p0 context.Context) ([]*AddrsResp, error) {
	return s.Internal.AddrList(p0)
}
func (s *IServiceStruct) AddrOperate(p0 context.Context, p1 *AddrsOperateReq) error {
	return s.Internal.AddrOperate(p0, p1)
}
func (s *IServiceStruct) ChainGetActor(p0 context.Context, p1 address.Address) (*types.Actor, error) {
	return s.Internal.ChainGetActor(p0, p1)
}
func (s *IServiceStruct) ChainGetHead(p0 context.Context) (*types.TipSet, error) {
	return s.Internal.ChainGetHead(p0)
}
func (s *IServiceStruct) ChainGetNetworkName(p0 context.Context) (types.NetworkName, error) {
	return s.Internal.ChainGetNetworkName(p0)
}
func (s *IServiceStruct) MinedBlockList(p0 context.Context, p1 MinedBlockListReq) (MinedBlockListResp, error) {
	return s.Internal.MinedBlockList(p0, p1)
}
func (s *IServiceStruct) MinerConfirmBeneficiary(p0 context.Context, p1 *MinerConfirmBeneficiaryReq) (address.Address, error) {
	return s.Internal.MinerConfirmBeneficiary(p0, p1)
}
func (s *IServiceStruct) MinerConfirmOwner(p0 context.Context, p1 *MinerSetOwnerReq) (address.Address, error) {
	return s.Internal.MinerConfirmOwner(p0, p1)
}
func (s *IServiceStruct) MinerConfirmWorker(p0 context.Context, p1 *MinerSetWorkerReq) error {
	return s.Internal.MinerConfirmWorker(p0, p1)
}
func (s *IServiceStruct) MinerCreate(p0 context.Context, p1 *MinerCreateReq) (address.Address, error) {
	return s.Internal.MinerCreate(p0, p1)
}
func (s *IServiceStruct) MinerGetDeadlines(p0 context.Context, p1 address.Address) (*dline.Info, error) {
	return s.Internal.MinerGetDeadlines(p0, p1)
}
func (s *IServiceStruct) MinerGetRetrievalAsk(p0 context.Context, p1 address.Address) (*retrievalmarket.Ask, error) {
	return s.Internal.MinerGetRetrievalAsk(p0, p1)
}
func (s *IServiceStruct) MinerGetStorageAsk(p0 context.Context, p1 address.Address) (*storagemarket.StorageAsk, error) {
	return s.Internal.MinerGetStorageAsk(p0, p1)
}
func (s *IServiceStruct) MinerInfo(p0 context.Context, p1 Address) (*MinerInfoResp, error) {
	return s.Internal.MinerInfo(p0, p1)
}
func (s *IServiceStruct) MinerList(p0 context.Context) ([]address.Address, error) {
	return s.Internal.MinerList(p0)
}
func (s *IServiceStruct) MinerSetBeneficiary(p0 context.Context, p1 *MinerSetBeneficiaryReq) (*types.PendingBeneficiaryChange, error) {
	return s.Internal.MinerSetBeneficiary(p0, p1)
}
func (s *IServiceStruct) MinerSetControllers(p0 context.Context, p1 *MinerSetControllersReq) ([]address.Address, error) {
	return s.Internal.MinerSetControllers(p0, p1)
}
func (s *IServiceStruct) MinerSetOwner(p0 context.Context, p1 *MinerSetOwnerReq) error {
	return s.Internal.MinerSetOwner(p0, p1)
}
func (s *IServiceStruct) MinerSetRetrievalAsk(p0 context.Context, p1 *MinerSetRetrievalAskReq) error {
	return s.Internal.MinerSetRetrievalAsk(p0, p1)
}
func (s *IServiceStruct) MinerSetStorageAsk(p0 context.Context, p1 *MinerSetAskReq) error {
	return s.Internal.MinerSetStorageAsk(p0, p1)
}
func (s *IServiceStruct) MinerSetWorker(p0 context.Context, p1 *MinerSetWorkerReq) (abi.ChainEpoch, error) {
	return s.Internal.MinerSetWorker(p0, p1)
}
func (s *IServiceStruct) MinerWinCount(p0 context.Context, p1 *MinerWinCountReq) (MinerWinCountResp, error) {
	return s.Internal.MinerWinCount(p0, p1)
}
func (s *IServiceStruct) MinerWithdrawFromMarket(p0 context.Context, p1 *MinerWithdrawBalanceReq) (abi.TokenAmount, error) {
	return s.Internal.MinerWithdrawFromMarket(p0, p1)
}
func (s *IServiceStruct) MinerWithdrawToBeneficiary(p0 context.Context, p1 *MinerWithdrawBalanceReq) (abi.TokenAmount, error) {
	return s.Internal.MinerWithdrawToBeneficiary(p0, p1)
}
func (s *IServiceStruct) Msg(p0 context.Context, p1 MsgID) (*MsgResp, error) {
	return s.Internal.Msg(p0, p1)
}
func (s *IServiceStruct) MsgDecodeParam2Json(p0 context.Context, p1 *MsgDecodeParamReq) ([]byte, error) {
	return s.Internal.MsgDecodeParam2Json(p0, p1)
}
func (s *IServiceStruct) MsgGetMethodName(p0 context.Context, p1 *MsgGetMethodNameReq) (string, error) {
	return s.Internal.MsgGetMethodName(p0, p1)
}
func (s *IServiceStruct) MsgMarkBad(p0 context.Context, p1 *MsgID) error {
	return s.Internal.MsgMarkBad(p0, p1)
}
func (s *IServiceStruct) MsgQuery(p0 context.Context, p1 *MsgQueryReq) ([]*MsgResp, error) {
	return s.Internal.MsgQuery(p0, p1)
}
func (s *IServiceStruct) MsgReplace(p0 context.Context, p1 *MsgReplaceReq) (cid.Cid, error) {
	return s.Internal.MsgReplace(p0, p1)
}
func (s *IServiceStruct) MsgSend(p0 context.Context, p1 *MsgSendReq) (string, error) {
	return s.Internal.MsgSend(p0, p1)
}
func (s *IServiceStruct) MsigAddSigner(p0 context.Context, p1 *MultisigChangeSignerReq) (*types.ProposeReturn, error) {
	return s.Internal.MsigAddSigner(p0, p1)
}
func (s *IServiceStruct) MsigApprove(p0 context.Context, p1 *MultisigApproveReq) (*types.ApproveReturn, error) {
	return s.Internal.MsigApprove(p0, p1)
}
func (s *IServiceStruct) MsigCancel(p0 context.Context, p1 *MultisigCancelReq) error {
	return s.Internal.MsigCancel(p0, p1)
}
func (s *IServiceStruct) MsigCreate(p0 context.Context, p1 *MultisigCreateReq) (address.Address, error) {
	return s.Internal.MsigCreate(p0, p1)
}
func (s *IServiceStruct) MsigInfo(p0 context.Context, p1 address.Address) (*types.MsigInfo, error) {
	return s.Internal.MsigInfo(p0, p1)
}
func (s *IServiceStruct) MsigListPropose(p0 context.Context, p1 address.Address) ([]*types.MsigTransaction, error) {
	return s.Internal.MsigListPropose(p0, p1)
}
func (s *IServiceStruct) MsigPropose(p0 context.Context, p1 *MultisigProposeReq) (*types.ProposeReturn, error) {
	return s.Internal.MsigPropose(p0, p1)
}
func (s *IServiceStruct) MsigRemoveSigner(p0 context.Context, p1 *MultisigChangeSignerReq) (*types.ProposeReturn, error) {
	return s.Internal.MsigRemoveSigner(p0, p1)
}
func (s *IServiceStruct) MsigSwapSigner(p0 context.Context, p1 *MultisigSwapSignerReq) (*types.ProposeReturn, error) {
	return s.Internal.MsigSwapSigner(p0, p1)
}
func (s *IServiceStruct) RetrievalDealList(p0 context.Context) ([]marketTypes.ProviderDealState, error) {
	return s.Internal.RetrievalDealList(p0)
}
func (s *IServiceStruct) Search(p0 context.Context, p1 SearchReq) (*SearchResp, error) {
	return s.Internal.Search(p0, p1)
}
func (s *IServiceStruct) SectorExtend(p0 context.Context, p1 SectorExtendReq) error {
	return s.Internal.SectorExtend(p0, p1)
}
func (s *IServiceStruct) SectorGet(p0 context.Context, p1 SectorGetReq) ([]*SectorResp, error) {
	return s.Internal.SectorGet(p0, p1)
}
func (s *IServiceStruct) SectorList(p0 context.Context, p1 SectorListReq) ([]*types.SectorOnChainInfo, error) {
	return s.Internal.SectorList(p0, p1)
}
func (s *IServiceStruct) SectorSum(p0 context.Context, p1 Address) (uint64, error) {
	return s.Internal.SectorSum(p0, p1)
}
func (s *IServiceStruct) StorageDeal(p0 context.Context, p1 Cid) (*marketTypes.MinerDeal, error) {
	return s.Internal.StorageDeal(p0, p1)
}
func (s *IServiceStruct) StorageDealList(p0 context.Context, p1 Address) ([]marketTypes.MinerDeal, error) {
	return s.Internal.StorageDealList(p0, p1)
}
func (s *IServiceStruct) StorageDealUpdateState(p0 context.Context, p1 StorageDealUpdateStateReq) error {
	return s.Internal.StorageDealUpdateState(p0, p1)
}
func (s *IServiceStruct) ThreadList(p0 context.Context) ([]*dep.ThreadInfo, error) {
	return s.Internal.ThreadList(p0)
}
func (s *IServiceStruct) ThreadStart(p0 context.Context, p1 *ThreadStartReq) error {
	return s.Internal.ThreadStart(p0, p1)
}
func (s *IServiceStruct) ThreadStop(p0 context.Context, p1 *ThreadStopReq) error {
	return s.Internal.ThreadStop(p0, p1)
}
func (s *IServiceStruct) WalletList(p0 context.Context) ([]address.Address, error) {
	return s.Internal.WalletList(p0)
}
func (s *IServiceStruct) WalletSignRecordQuery(p0 context.Context, p1 *WalletSignRecordQueryReq) ([]WalletSignRecordResp, error) {
	return s.Internal.WalletSignRecordQuery(p0, p1)
}
