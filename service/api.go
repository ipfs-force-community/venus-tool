package service

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/dline"
	"github.com/filecoin-project/venus/venus-shared/types"
	marketTypes "github.com/filecoin-project/venus/venus-shared/types/market"
	"github.com/ipfs-force-community/venus-tool/dep"
	"github.com/ipfs/go-cid"
)

//go:generate go run ../utils/gen/api-gen.go

type IService interface {
	ChainGetHead(ctx context.Context) (*types.TipSet, error)                       // GET:/chain/head
	ChainGetActor(ctx context.Context, addr address.Address) (*types.Actor, error) // GET:/chain/actor
	ChainGetNetworkName(ctx context.Context) (types.NetworkName, error)            // GET:/chain/networkname

	MsgSend(ctx context.Context, params *MsgSendReq) (string, error)                 // POST:/msg/send
	MsgQuery(ctx context.Context, params *MsgQueryReq) ([]*MsgResp, error)           // GET:/msg/query
	Msg(ctx context.Context, id MsgID) (*MsgResp, error)                             // GET:/msg/:ID
	MsgReplace(ctx context.Context, params *MsgReplaceReq) (cid.Cid, error)          // POST:/msg/replace
	MsgDecodeParam2Json(ctx context.Context, req *MsgDecodeParamReq) ([]byte, error) // POST:/msg/decodeparam
	MsgGetMethodName(ctx context.Context, req *MsgGetMethodNameReq) (string, error)  // GET:/msg/getmethodname
	MsgMarkBad(ctx context.Context, req *MsgID) error                                // POST:/msg/markbad/:ID

	AddrOperate(ctx context.Context, params *AddrsOperateReq) error // PUT:/addr/operate
	AddrInfo(ctx context.Context, addr Address) (*AddrsResp, error) // GET:/addr/info/:Address
	// return the addr setting from messager
	AddrList(ctx context.Context) ([]*AddrsResp, error) // GET:/addr/list
	// return addr registered in wallet
	WalletList(ctx context.Context) ([]address.Address, error)                                                // GET:/wallet/list
	WalletSignRecordQuery(ctx context.Context, req *WalletSignRecordQueryReq) ([]WalletSignRecordResp, error) // GET:/wallet/signrecord

	MinerInfo(ctx context.Context, mAddr Address) (*MinerInfoResp, error)                                                // GET:/miner/info/:Address
	MinerList(ctx context.Context) ([]address.Address, error)                                                            // GET:/miner/list
	MinerCreate(ctx context.Context, params *MinerCreateReq) (address.Address, error)                                    // POST:/miner/create
	MinerGetStorageAsk(ctx context.Context, mAddr address.Address) (*storagemarket.StorageAsk, error)                    // GET:/miner/storageask
	MinerGetRetrievalAsk(ctx context.Context, mAddr address.Address) (*retrievalmarket.Ask, error)                       // GET:/miner/retrievalask
	MinerSetStorageAsk(ctx context.Context, p *MinerSetAskReq) error                                                     // PUT:/miner/storageask
	MinerSetRetrievalAsk(ctx context.Context, p *MinerSetRetrievalAskReq) error                                          // PUT:/miner/retrievalask
	MinerGetDeadlines(ctx context.Context, mAddr address.Address) (*dline.Info, error)                                   // GET:/miner/deadline
	MinerSetOwner(ctx context.Context, p *MinerSetOwnerReq) error                                                        // PUT:/miner/owner
	MinerConfirmOwner(ctx context.Context, p *MinerSetOwnerReq) (oldOwner address.Address, err error)                    // PUT:/miner/confirmowner
	MinerSetWorker(ctx context.Context, req *MinerSetWorkerReq) (WorkerChangeEpoch abi.ChainEpoch, err error)            // PUT:/miner/worker
	MinerConfirmWorker(ctx context.Context, req *MinerSetWorkerReq) error                                                // PUT:/miner/confirmworker
	MinerSetControllers(ctx context.Context, req *MinerSetControllersReq) (oldController []address.Address, err error)   // PUT:/miner/controllers
	MinerSetBeneficiary(ctx context.Context, req *MinerSetBeneficiaryReq) (*types.PendingBeneficiaryChange, error)       // PUT:/miner/beneficiary
	MinerConfirmBeneficiary(ctx context.Context, req *MinerConfirmBeneficiaryReq) (confirmor address.Address, err error) // PUT:/miner/confirmbeneficiary
	// MinerWithdrawFromMarket withdraws funds from miner to it's beneficiary
	MinerWithdrawToBeneficiary(ctx context.Context, req *MinerWithdrawBalanceReq) (abi.TokenAmount, error) // PUT:/miner/withdrawbeneficiary
	// MinerWithdrawFromMarket withdraw balance from market to miner's owner or worker
	MinerWithdrawFromMarket(ctx context.Context, req *MinerWithdrawBalanceReq) (abi.TokenAmount, error) // PUT:/miner/withdrawmarket
	MinerWinCount(ctx context.Context, req *MinerWinCountReq) (MinerWinCountResp, error)                // GET:/miner/wincount

	StorageDealList(ctx context.Context, miner Address) ([]marketTypes.MinerDeal, error) // GET:/deal/storage/:Address
	StorageDeal(ctx context.Context, proposalCid Cid) (*marketTypes.MinerDeal, error)    // GET:/deal/storage/info/:Cid
	StorageDealUpdateState(ctx context.Context, req StorageDealUpdateStateReq) error     // PUT:/deal/storage/state
	RetrievalDealList(ctx context.Context) ([]marketTypes.ProviderDealState, error)      // GET:/deal/retrieval

	SectorExtend(ctx context.Context, req SectorExtendReq) error                           // PUT:/sector/extend
	SectorGet(ctx context.Context, req SectorGetReq) ([]*SectorResp, error)                // GET:/sector/get
	SectorList(ctx context.Context, req SectorListReq) ([]*types.SectorOnChainInfo, error) // GET:/sector/list
	SectorSum(ctx context.Context, miner Address) (uint64, error)                          // GET:/sector/sum

	MsigCreate(ctx context.Context, req *MultisigCreateReq) (address.Address, error)                  // POST:/msig/create
	MsigInfo(ctx context.Context, msig address.Address) (*types.MsigInfo, error)                      // GET:/msig/info
	MsigPropose(ctx context.Context, req *MultisigProposeReq) (*types.ProposeReturn, error)           // POST:/msig/propose
	MsigListPropose(ctx context.Context, msig address.Address) ([]*types.MsigTransaction, error)      // GET:/msig/proposes
	MsigAddSigner(ctx context.Context, req *MultisigChangeSignerReq) (*types.ProposeReturn, error)    // POST:/msig/signer/ass
	MsigRemoveSigner(ctx context.Context, req *MultisigChangeSignerReq) (*types.ProposeReturn, error) // POST:/msig/signer/remove
	MsigApprove(ctx context.Context, req *MultisigApproveReq) (*types.ApproveReturn, error)           // POST:/msig/approve
	MsigCancel(ctx context.Context, req *MultisigCancelReq) error                                     // POST:/msig/cancel
	MsigSwapSigner(ctx context.Context, req *MultisigSwapSignerReq) (*types.ProposeReturn, error)     // POST:/msig/signer/swap

	ThreadList(ctx context.Context) ([]*dep.ThreadInfo, error)  // GET:/thread/list
	ThreadStop(ctx context.Context, req *ThreadStopReq) error   // PUT:/thread/stop
	ThreadStart(ctx context.Context, req *ThreadStartReq) error // PUT:/thread/start

	Search(ctx context.Context, req SearchReq) (*SearchResp, error)                        // GET:/search/:Key
	MinedBlockList(ctx context.Context, req MinedBlockListReq) (MinedBlockListResp, error) // GET:/minedblock/list
}
