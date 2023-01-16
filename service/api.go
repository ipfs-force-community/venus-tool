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
	"github.com/ipfs/go-cid"
)

//go:generate go run ../utils/gen/api-gen.go

type IService interface {
	ChainGetHead(ctx context.Context) (*types.TipSet, error)                       // GET:/chain/head
	ChainGetActor(ctx context.Context, addr address.Address) (*types.Actor, error) // GET:/chain/actor
	ChainGetNetworkName(ctx context.Context) (types.NetworkName, error)            // GET:/chain/networkname

	MsgSend(ctx context.Context, params *MsgSendReq) (string, error)        // POST:/msg/send
	MsgQuery(ctx context.Context, params *MsgQueryReq) ([]*MsgResp, error)  // GET:/msg/query
	MsgReplace(ctx context.Context, params *MsgReplaceReq) (cid.Cid, error) // POST:/msg/replace

	AddrOperate(ctx context.Context, params *AddrsOperateReq) error // PUT:/addr/operate
	AddrList(ctx context.Context) ([]*AddrsResp, error)             // GET:/addr/list

	MinerInfo(ctx context.Context, mAddr address.Address) (*MinerInfoResp, error)                                        // GET:/miner/info
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

	StorageDealList(ctx context.Context, miner address.Address) ([]marketTypes.MinerDeal, error) // GET:/deal/storage
	StorageDealUpdateState(ctx context.Context, req StorageDealUpdateStateReq) error             // PUT:/deal/storage/state
	RetrievalDealList(ctx context.Context) ([]marketTypes.ProviderDealState, error)              // GET:/deal/retrieval

	SectorExtend(ctx context.Context, req SectorExtendReq) error            // PUT:/sector/extend
	SectorGet(ctx context.Context, req SectorGetReq) ([]*SectorResp, error) // GET:/sector/get

	MsigCreate(ctx context.Context, req *MultiSigCreateReq) (address.Address, error) // POST:/msig/create
}
