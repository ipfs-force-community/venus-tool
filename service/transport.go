package service

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/dline"
	power2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/power"
	lminer "github.com/filecoin-project/venus/venus-shared/actors/builtin/miner"
	"github.com/filecoin-project/venus/venus-shared/types"
	marketTypes "github.com/filecoin-project/venus/venus-shared/types/market"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs-force-community/venus-tool/utils"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type MsgResp struct {
	msgTypes.Message
	MethodName string
}

func (mr *MsgResp) MarshalJSON() ([]byte, error) {
	type Msg msgTypes.Message
	type temp struct {
		Msg
		MethodName string
	}
	return json.Marshal(temp{
		Msg:        Msg(mr.Message),
		MethodName: mr.MethodName,
	})
}

type MsgReplaceReq = msgTypes.ReplacMessageParams

type MsgQueryReq struct {
	msgTypes.MsgQueryParams
	IsFailed    bool
	IsBlocked   bool
	BlockedTime time.Duration
	ID          string
	Nonce       uint64
}

type MsgSendReq struct {
	From   address.Address
	To     address.Address
	Value  abi.TokenAmount
	Method abi.MethodNum
	Params EncodedParams

	msgTypes.SendSpec
}

type EncodingType string

type EncodedParams struct {
	Data    []byte
	EncType EncodingType
}

const (
	EncNull EncodingType = ""
	EncHex  EncodingType = "hex"
	EncJson EncodingType = "json"
)

func (ep *EncodedParams) DecodeJSON(actorCode cid.Cid, method abi.MethodNum) (out []byte, err error) {
	methodMeta, err := utils.GetMethodMeta(actorCode, method)
	if err != nil {
		return nil, err
	}

	p := reflect.New(methodMeta.Params.Elem()).Interface().(cbg.CBORMarshaler)
	if err := json.Unmarshal(ep.Data, p); err != nil {
		return nil, fmt.Errorf("unmarshaling input into params type: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := p.MarshalCBOR(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ep *EncodedParams) DecodeHex() (out []byte, err error) {
	return hex.DecodeString(string(ep.Data))
}

type MsgGetMethodNameReq struct {
	To     address.Address
	Method abi.MethodNum
}

type MsgDecodeParamReq struct {
	To     address.Address
	Method abi.MethodNum
	Params []byte
}

type AddrOperateType string

var (
	DeleteAddress    AddrOperateType = "delete"
	ActiveAddress    AddrOperateType = "active"
	ForbiddenAddress AddrOperateType = "forbidden"
	SetAddress       AddrOperateType = "set"
)

type AddrsResp msgTypes.Address

type AddrsOperateReq struct {
	msgTypes.AddressSpec
	Operate      AddrOperateType
	SelectMsgNum uint64
	IsSetSpec    bool
}

type MinerSetAskReq struct {
	Miner         address.Address
	Price         abi.TokenAmount
	VerifiedPrice abi.TokenAmount
	Duration      abi.ChainEpoch
	MinPieceSize  abi.PaddedPieceSize
	MaxPieceSize  abi.PaddedPieceSize
}

type MinerSetRetrievalAskReq struct {
	retrievalmarket.Ask
	Miner address.Address
}

type MinerSetBeneficiaryReq struct {
	Miner address.Address
	types.ChangeBeneficiaryParams
}

type MinerConfirmBeneficiaryReq struct {
	Miner          address.Address
	NewBeneficiary address.Address
	ByNominee      bool
}

type StorageDealUpdateStateReq struct {
	ProposalCid cid.Cid
	State       storagemarket.StorageDealStatus
	PieceStatus marketTypes.PieceStatus
}

type MinerCreateReq struct {
	power2.CreateMinerParams
	From       address.Address
	SectorSize abi.SectorSize
	MsgId      string
}

type MinerInfoResp struct {
	types.MinerInfo
	types.MinerPower
	AvailBalance abi.TokenAmount
	Deadline     dline.Info
}

type MinerSetOwnerReq struct {
	Miner    address.Address
	NewOwner address.Address
}

type MinerSetWorkerReq struct {
	Miner     address.Address
	NewWorker address.Address
}

type MinerSetControllersReq struct {
	Miner          address.Address
	NewControllers []address.Address
}

type MinerWithdrawBalanceReq struct {
	Miner  address.Address
	To     address.Address
	Amount abi.TokenAmount
}

type SectorExtendReq struct {
	Miner         address.Address
	SectorNumbers []abi.SectorNumber
	Expiration    abi.ChainEpoch
}

type SectorGetReq struct {
	Miner         address.Address
	SectorNumbers []abi.SectorNumber
}

type SectorResp struct {
	types.SectorOnChainInfo
	SectorLocation lminer.SectorLocation
}

type MultisigCreateReq struct {
	From               address.Address
	Signers            []address.Address
	ApprovalsThreshold uint64
	LockedDuration     abi.ChainEpoch
	Value              abi.TokenAmount
}

type MultisigProposeReq struct {
	Msig   address.Address
	From   address.Address
	To     address.Address
	Value  abi.TokenAmount
	Method abi.MethodNum
	Params EncodedParams
}

type MultisigAddSignerReq struct {
	NewSigner         address.Address
	Proposer          address.Address
	Msig              address.Address
	IncreaseThresHold bool
}

type MultisigTransactionReq struct {
	Msig     address.Address
	Proposer address.Address
	TxID     uint64
}

type MultisigApproveReq = MultisigTransactionReq
type MultisigCancelReq = MultisigTransactionReq
