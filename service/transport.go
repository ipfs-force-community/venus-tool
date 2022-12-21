package service

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	nodeV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs-force-community/venus-tool/utils"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type MsgResp struct {
	msgTypes.Message
	MethodName string
}

func (mr *MsgResp) getMethodName(node nodeV1.IActor) (string, error) {
	methodMeta, err := utils.GetMethodMeta(node, mr.To, mr.Method)
	if err != nil {
		return "", err
	}
	mr.MethodName = methodMeta.Name
	return methodMeta.Name, nil
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

type MsgQueryReq struct {
	msgTypes.MsgQueryParams
	IsFailed    bool
	IsBlocked   bool
	BlockedTime time.Duration
	ID          string
	Nonce       uint64
}

type MsgSendReq struct {
	From    address.Address
	To      address.Address
	Value   abi.TokenAmount
	Method  abi.MethodNum
	Params  []byte
	EncType EncodingType

	msgTypes.SendSpec
}

type MsgReplaceReq = msgTypes.ReplacMessageParams

type EncodingType string

const (
	EncNull EncodingType = ""
	EncHex  EncodingType = "hex"
	EncJson EncodingType = "json"
)

func (sp *MsgSendReq) Decode(node nodeV1.IActor) (params []byte, err error) {

	switch sp.EncType {
	case EncNull:
		params = sp.Params
	case EncJson:
		params, err = sp.DecodeJSON(node)
		if err != nil {
			return nil, fmt.Errorf("failed to decode json params: %w", err)
		}
	case EncHex:
		params, err = sp.DecodeHex()
		if err != nil {
			return nil, fmt.Errorf("failed to decode hex params: %w", err)
		}
	default:
		return nil, fmt.Errorf("unexpected param type %s", sp.EncType)
	}

	return params, nil
}

func (sp *MsgSendReq) DecodeJSON(node nodeV1.IActor) (out []byte, err error) {
	methodMeta, err := utils.GetMethodMeta(node, sp.To, sp.Method)
	if err != nil {
		return nil, err
	}

	p := reflect.New(methodMeta.Params.Elem()).Interface().(cbg.CBORMarshaler)
	if err := json.Unmarshal(sp.Params, p); err != nil {
		return nil, fmt.Errorf("unmarshaling input into params type: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := p.MarshalCBOR(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (sp *MsgSendReq) DecodeHex() (out []byte, err error) {
	return hex.DecodeString(string(sp.Params))
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
