package service

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	nodeV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	msgTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/filecoin-project/venus/venus-shared/utils"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type SendParams struct {
	From    address.Address
	To      address.Address
	Value   abi.TokenAmount
	Method  abi.MethodNum
	Params  []byte
	EncType EncodingType

	msgTypes.SendSpec
}

type EncodingType string

const (
	EncNull EncodingType = ""
	EncHex  EncodingType = "hex"
	EncJson EncodingType = "json"
)

func (sp *SendParams) Decode(node nodeV1.IActor) (params []byte, err error) {

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

func (sp *SendParams) DecodeJSON(node nodeV1.IActor) (out []byte, err error) {
	methodMeta, err := getMethodMeta(node, sp.To, sp.Method)
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

func (sp *SendParams) DecodeHex() (out []byte, err error) {
	return hex.DecodeString(string(sp.Params))
}

func getMethodMeta(node nodeV1.IActor, to address.Address, method abi.MethodNum) (utils.MethodMeta, error) {
	ctx := context.Background()
	act, err := node.StateGetActor(ctx, to, venusTypes.EmptyTSK)
	if err != nil {
		return utils.MethodMeta{}, err
	}

	methodMeta, found := utils.MethodsMap[act.Code][method]
	if !found {
		return utils.MethodMeta{}, fmt.Errorf("method %d not found on actor %s", method, act.Code)
	}
	return methodMeta, nil
}
