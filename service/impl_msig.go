package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	init2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/init"
	"github.com/filecoin-project/venus/venus-shared/types"
)

func (s *ServiceImpl) MsigCreate(ctx context.Context, req *MultisigCreateReq) (address.Address, error) {
	var err error
	// check params
	if req.ApprovalsThreshold < 1 {
		return address.Undef, fmt.Errorf("threshold(%d) must be greater than 1", req.ApprovalsThreshold)
	}

	if uint64(len(req.Signers)) < req.ApprovalsThreshold {
		return address.Undef, fmt.Errorf("signers(%d) must be greater than threshold(%d)", len(req.Signers), req.ApprovalsThreshold)
	}

	if req.Value.LessThan(big.Zero()) {
		return address.Undef, fmt.Errorf("value(%s) must be equal or greater than 0", req.Value)
	}

	if req.LockedDuration < 0 {
		return address.Undef, fmt.Errorf("unlockAt(%d) must be equal or greater than 0", req.LockedDuration)
	}

	// check signers
	set := make(map[address.Address]struct{})
	for _, signer := range req.Signers {
		id, err := s.Node.StateLookupID(ctx, signer, types.EmptyTSK)
		if err != nil {
			return address.Undef, fmt.Errorf("lookup signer(%s) failed: %s", signer, err)
		}
		if _, ok := set[id]; ok {
			return address.Undef, fmt.Errorf("duplicate signer(%s)", signer)
		} else {
			set[id] = struct{}{}
		}
	}

	msgPrototype, err := s.Multisig.MsigCreate(ctx, req.ApprovalsThreshold, req.Signers, req.LockedDuration, req.Value, req.From, big.Zero())
	if err != nil {
		return address.Undef, fmt.Errorf("create multisig Prototype failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &msgPrototype.Message, nil)
	if err != nil {
		return address.Undef, fmt.Errorf("push message failed: %s", err)
	}

	var execRet init2.ExecReturn
	if err := execRet.UnmarshalCBOR(bytes.NewReader(msg.Receipt.Return)); err != nil {
		return address.Undef, fmt.Errorf("unmarshal multisig create exec return failed: %s", err)
	}

	return execRet.RobustAddress, nil
}

func (s *ServiceImpl) MsigInfo(ctx context.Context, msig address.Address) (*types.MsigInfo, error) {
	info, err := s.Multisig.StateMsigInfo(ctx, msig, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get multisig info failed: %s", err)
	}

	return info, nil
}

func (s *ServiceImpl) MsigPropose(ctx context.Context, req *MultisigProposeReq) (*types.ProposeReturn, error) {
	var err error

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
		case EncNull:
			return req.Data, nil
		default:
			return nil, fmt.Errorf("unknown encoding type: %s", req.EncType)
		}
	}

	if req.Value.LessThan(big.Zero()) {
		return nil, fmt.Errorf("value(%s) must be equal or greater than 0", req.Value)
	}

	params, err := dec(req.Params, req.To, req.Method)
	if err != nil {
		return nil, fmt.Errorf("decode params failed: %s", err)
	}

	msgPrototype, err := s.Multisig.MsigPropose(ctx, req.Msig, req.To, req.Value, req.From, uint64(req.Method), params)
	if err != nil {
		return nil, fmt.Errorf("create multisig propose Prototype failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &msgPrototype.Message, nil)
	if err != nil {
		return nil, fmt.Errorf("push message failed: %s", err)
	}

	var msgReturn types.ProposeReturn
	err = msgReturn.UnmarshalCBOR(bytes.NewReader(msg.Receipt.Return))
	if err != nil {
		return nil, fmt.Errorf("unmarshal propose return failed: %s", err)
	}

	return &msgReturn, nil
}

func (s *ServiceImpl) MsigListPropose(ctx context.Context, msig address.Address) ([]*types.MsigTransaction, error) {
	var err error

	ret, err := s.Multisig.MsigGetPending(ctx, msig, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("create multisig get pending Prototype failed: %s", err)
	}

	return ret, nil
}

func (s *ServiceImpl) MsigAddSigner(ctx context.Context, req *MultisigChangeSignerReq) (*types.ProposeReturn, error) {
	var err error

	_, err = s.Node.StateLookupID(ctx, req.NewSigner, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("lookup signer(%s) failed: %s", req.NewSigner, err)
	}

	msgPrototype, err := s.Multisig.MsigAddPropose(ctx, req.Msig, req.Proposer, req.NewSigner, req.AlterThresHold)
	if err != nil {
		return nil, fmt.Errorf("create multisig add propose Prototype failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &msgPrototype.Message, nil)
	if err != nil {
		return nil, fmt.Errorf("push message failed: %s", err)
	}

	var msgReturn types.ProposeReturn
	err = msgReturn.UnmarshalCBOR(bytes.NewReader(msg.Receipt.Return))
	if err != nil {
		return nil, fmt.Errorf("unmarshal add propose return failed: %s", err)
	}

	return &msgReturn, nil
}

func (s *ServiceImpl) MsigRemoveSigner(ctx context.Context, req *MultisigChangeSignerReq) (*types.ProposeReturn, error) {
	var err error

	_, err = s.Node.StateLookupID(ctx, req.NewSigner, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("lookup signer(%s) failed: %s", req.NewSigner, err)
	}

	msgPrototype, err := s.Multisig.MsigRemoveSigner(ctx, req.Msig, req.Proposer, req.NewSigner, req.AlterThresHold)
	if err != nil {
		return nil, fmt.Errorf("create multisig remove propose Prototype failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &msgPrototype.Message, nil)
	if err != nil {
		return nil, fmt.Errorf("push message failed: %s", err)
	}

	var msgReturn types.ProposeReturn
	err = msgReturn.UnmarshalCBOR(bytes.NewReader(msg.Receipt.Return))
	if err != nil {
		return nil, fmt.Errorf("unmarshal remove propose return failed: %s", err)
	}

	return &msgReturn, nil
}

func (s *ServiceImpl) MsigSwapSigner(ctx context.Context, req *MultisigSwapSignerReq) (*types.ProposeReturn, error) {
	var err error

	_, err = s.Node.StateLookupID(ctx, req.Proposer, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("lookup from(%s) failed: %s", req.Proposer, err)
	}

	msgPrototype, err := s.Multisig.MsigSwapPropose(ctx, req.Msig, req.Proposer, req.OldSigner, req.NewSigner)
	if err != nil {
		return nil, fmt.Errorf("create multisig swap propose Prototype failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &msgPrototype.Message, nil)
	if err != nil {
		return nil, fmt.Errorf("push message failed: %s", err)
	}

	var msgReturn types.ProposeReturn
	err = msgReturn.UnmarshalCBOR(bytes.NewReader(msg.Receipt.Return))
	if err != nil {
		return nil, fmt.Errorf("unmarshal swap propose return failed: %s", err)
	}

	return &msgReturn, nil
}

func (s *ServiceImpl) MsigApprove(ctx context.Context, req *MultisigApproveReq) (*types.ApproveReturn, error) {
	var err error

	_, err = s.Node.StateLookupID(ctx, req.Proposer, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("lookup proposer(%s) failed: %s", req.Proposer, err)
	}

	msgPrototype, err := s.Multisig.MsigApprove(ctx, req.Msig, req.TxID, req.Proposer)
	if err != nil {
		return nil, fmt.Errorf("create multisig approve Prototype failed: %s", err)
	}

	msg, err := s.PushMessageAndWait(ctx, &msgPrototype.Message, nil)
	if err != nil {
		return nil, fmt.Errorf("push message failed: %s", err)
	}

	var msgReturn types.ApproveReturn
	err = msgReturn.UnmarshalCBOR(bytes.NewReader(msg.Receipt.Return))
	if err != nil {
		return nil, fmt.Errorf("unmarshal approve return failed: %s", err)
	}

	return &msgReturn, nil
}

func (s *ServiceImpl) MsigCancel(ctx context.Context, req *MultisigCancelReq) error {
	var err error

	_, err = s.Node.StateLookupID(ctx, req.Proposer, types.EmptyTSK)
	if err != nil {
		return fmt.Errorf("lookup proposer(%s) failed: %s", req.Proposer, err)
	}

	msgPrototype, err := s.Multisig.MsigCancel(ctx, req.Msig, req.TxID, req.Proposer)
	if err != nil {
		return fmt.Errorf("create multisig cancel Prototype failed: %s", err)
	}

	_, err = s.PushMessageAndWait(ctx, &msgPrototype.Message, nil)
	if err != nil {
		return fmt.Errorf("push message failed: %s", err)
	}

	return nil
}
