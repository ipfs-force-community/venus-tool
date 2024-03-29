package multisig

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	actorstypes "github.com/filecoin-project/go-state-types/actors"
	"github.com/filecoin-project/go-state-types/big"
	multisig2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/multisig"

	"github.com/filecoin-project/venus/venus-shared/actors"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin/multisig"
	v1api "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/types"

	"github.com/filecoin-project/venus/venus-shared/actors/adt"
	"github.com/filecoin-project/venus/venus-shared/blockstore"
	cbor "github.com/ipfs/go-ipld-cbor"
)

var _ IMultiSig = &multiSig{}

type multiSig struct {
	state      v1api.IChain
	mpool      v1api.IMessagePool
	blockstore v1api.IBlockStore
}

type MsigProposeResponse int

const (
	MsigApprove MsigProposeResponse = iota
	MsigCancel
)

func NewMultiSig(m v1api.FullNode) IMultiSig {
	return &multiSig{
		state:      m,
		mpool:      m,
		blockstore: m,
	}
}

func (a *multiSig) Store(ctx context.Context) adt.Store {
	return adt.WrapStore(ctx, cbor.NewCborStore(blockstore.NewAPIBlockstore(a.blockstore)))
}

func (a *multiSig) messageBuilder(ctx context.Context, from address.Address) (multisig.MessageBuilder, error) {
	nver, err := a.state.StateNetworkVersion(ctx, types.EmptyTSK)
	if err != nil {
		return nil, err
	}
	aver, err := actorstypes.VersionForNetwork(nver)
	if err != nil {
		return nil, err
	}
	return multisig.Message(aver, from), nil
}

// MsigCreate creates a multisig wallet
// It takes the following params: <required number of senders>, <approving addresses>, <unlock duration>
// <initial balance>, <sender address of the create msg>, <gas price>
func (a *multiSig) MsigCreate(ctx context.Context, req uint64, addrs []address.Address, duration abi.ChainEpoch, val types.BigInt, src address.Address, gp types.BigInt) (*types.MessagePrototype, error) {
	mb, err := a.messageBuilder(ctx, src)
	if err != nil {
		return nil, err
	}

	msg, err := mb.Create(addrs, req, 0, duration, val)
	if err != nil {
		return nil, err
	}

	return &types.MessagePrototype{
		Message:    *msg,
		ValidNonce: false,
	}, nil
}

func (a *multiSig) StateMsigInfo(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*types.MsigInfo, error) {
	ret := &types.MsigInfo{}

	ts, err := a.state.ChainGetTipSet(ctx, tsk)
	if err != nil {
		return nil, err
	}

	act, err := a.state.StateGetActor(ctx, addr, tsk)
	if err != nil {
		return nil, fmt.Errorf("failed to load multisig actor: %w", err)
	}
	msas, err := multisig.Load(a.Store(ctx), act)
	if err != nil {
		return nil, fmt.Errorf("failed to load multisig actor state: %w", err)
	}

	ret.ApprovalsThreshold, err = msas.Threshold()
	if err != nil {
		return nil, fmt.Errorf("failed to get threshold: %w", err)
	}

	ret.Signers, err = msas.Signers()
	if err != nil {
		return nil, fmt.Errorf("failed to get signers: %w", err)
	}

	ret.InitialBalance, err = msas.InitialBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get initial balance: %w", err)
	}

	ret.CurrentBalance = act.Balance
	ret.LockBalance, err = msas.LockedBalance(ts.Height())
	if err != nil {
		return nil, fmt.Errorf("failed to get locked balance: %w", err)
	}

	ret.StartEpoch, err = msas.StartEpoch()
	if err != nil {
		return nil, fmt.Errorf("failed to get start epoch: %w", err)
	}

	ret.UnlockDuration, err = msas.UnlockDuration()
	if err != nil {
		return nil, fmt.Errorf("failed to get unlocked duration %w", err)
	}

	return ret, nil
}

func (a *multiSig) MsigPropose(ctx context.Context, msig address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.MessagePrototype, error) {
	mb, err := a.messageBuilder(ctx, src)
	if err != nil {
		return nil, err
	}

	msg, err := mb.Propose(msig, to, amt, abi.MethodNum(method), params)
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal: %w", err)
	}

	return &types.MessagePrototype{
		Message:    *msg,
		ValidNonce: false,
	}, nil
}

func (a *multiSig) MsigAddPropose(ctx context.Context, msig address.Address, src address.Address, newAdd address.Address, inc bool) (*types.MessagePrototype, error) {
	enc, actErr := serializeAddParams(newAdd, inc)
	if actErr != nil {
		return nil, actErr
	}

	return a.MsigPropose(ctx, msig, msig, big.Zero(), src, uint64(multisig.Methods.AddSigner), enc)
}

func (a *multiSig) MsigAddApprove(ctx context.Context, msig address.Address, src address.Address, txID uint64, proposer address.Address, newAdd address.Address, inc bool) (*types.MessagePrototype, error) {
	enc, actErr := serializeAddParams(newAdd, inc)
	if actErr != nil {
		return nil, actErr
	}

	return a.MsigApproveTxnHash(ctx, msig, txID, proposer, msig, big.Zero(), src, uint64(multisig.Methods.AddSigner), enc)
}

// MsigAddApprove approves a previously proposed AddSigner message
// It takes the following params: <multisig address>, <sender address of the approve msg>, <proposed message ID>,
// <proposer address>, <new signer>, <whether the number of required signers should be increased>
func (a *multiSig) MsigAddCancel(ctx context.Context, msig address.Address, src address.Address, txID uint64, newAdd address.Address, inc bool) (*types.MessagePrototype, error) {
	enc, actErr := serializeAddParams(newAdd, inc)
	if actErr != nil {
		return nil, actErr
	}

	return a.MsigCancelTxnHash(ctx, msig, txID, msig, big.Zero(), src, uint64(multisig.Methods.AddSigner), enc)
}

func (a *multiSig) MsigCancelTxnHash(ctx context.Context, msig address.Address, txID uint64, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.MessagePrototype, error) {
	return a.msigApproveOrCancelTxnHash(ctx, MsigCancel, msig, txID, src, to, amt, src, method, params)
}

// MsigSwapPropose proposes swapping 2 signers in the multisig
// It takes the following params: <multisig address>, <sender address of the propose msg>,
// <old signer>, <new signer>
func (a *multiSig) MsigSwapPropose(ctx context.Context, msig address.Address, src address.Address, oldAdd address.Address, newAdd address.Address) (*types.MessagePrototype, error) {
	enc, actErr := serializeSwapParams(oldAdd, newAdd)
	if actErr != nil {
		return nil, actErr
	}

	return a.MsigPropose(ctx, msig, msig, big.Zero(), src, uint64(multisig.Methods.SwapSigner), enc)
}

// MsigSwapApprove approves a previously proposed SwapSigner
// It takes the following params: <multisig address>, <sender address of the approve msg>, <proposed message ID>,
// <proposer address>, <old signer>, <new signer>
func (a *multiSig) MsigSwapApprove(ctx context.Context, msig address.Address, src address.Address, txID uint64, proposer address.Address, oldAdd address.Address, newAdd address.Address) (*types.MessagePrototype, error) {
	enc, actErr := serializeSwapParams(oldAdd, newAdd)
	if actErr != nil {
		return nil, actErr
	}

	return a.MsigApproveTxnHash(ctx, msig, txID, proposer, msig, big.Zero(), src, uint64(multisig.Methods.SwapSigner), enc)
}

func (a *multiSig) MsigSwapCancel(ctx context.Context, msig address.Address, src address.Address, txID uint64, oldAdd address.Address, newAdd address.Address) (*types.MessagePrototype, error) {
	enc, actErr := serializeSwapParams(oldAdd, newAdd)
	if actErr != nil {
		return nil, actErr
	}

	return a.MsigCancelTxnHash(ctx, msig, txID, msig, big.Zero(), src, uint64(multisig.Methods.SwapSigner), enc)
}

// MsigSwapCancel cancels a previously proposed SwapSigner message
// It takes the following params: <multisig address>, <sender address of the cancel msg>, <proposed message ID>,
// <old signer>, <new signer>
func (a *multiSig) MsigApprove(ctx context.Context, msig address.Address, txID uint64, src address.Address) (*types.MessagePrototype, error) {
	return a.msigApproveOrCancelSimple(ctx, MsigApprove, msig, txID, src)
}

// MsigApproveTxnHash approves a previously-proposed multisig message, specified
// using both transaction ID and a hash of the parameters used in the
// proposal. This method of approval can be used to ensure you only approve
// exactly the transaction you think you are.
// It takes the following params: <multisig address>, <proposed message ID>, <proposer address>, <recipient address>, <value to transfer>,
// <sender address of the approve msg>, <method to call in the proposed message>, <params to include in the proposed message>
func (a *multiSig) MsigApproveTxnHash(ctx context.Context, msig address.Address, txID uint64, proposer address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.MessagePrototype, error) {
	return a.msigApproveOrCancelTxnHash(ctx, MsigApprove, msig, txID, proposer, to, amt, src, method, params)
}

// MsigCancel cancels a previously-proposed multisig message
// It takes the following params: <multisig address>, <proposed transaction ID>, <recipient address>, <value to transfer>,
// <sender address of the cancel msg>, <method to call in the proposed message>, <params to include in the proposed message>
func (a *multiSig) MsigCancel(ctx context.Context, msig address.Address, txID uint64, src address.Address) (*types.MessagePrototype, error) {
	return a.msigApproveOrCancelSimple(ctx, MsigCancel, msig, txID, src)
}

// MsigRemoveSigner proposes the removal of a signer from the multisig.
// It accepts the multisig to make the change on, the proposer address to
// send the message from, the address to be removed, and a boolean
// indicating whether or not the signing threshold should be lowered by one
// along with the address removal.
func (a *multiSig) MsigRemoveSigner(ctx context.Context, msig address.Address, proposer address.Address, toRemove address.Address, decrease bool) (*types.MessagePrototype, error) {
	enc, actErr := serializeRemoveParams(toRemove, decrease)
	if actErr != nil {
		return nil, actErr
	}

	return a.MsigPropose(ctx, msig, msig, types.NewInt(0), proposer, uint64(multisig.Methods.RemoveSigner), enc)
}

// MsigGetVested returns the amount of FIL that vested in a multisig in a certain period.
// It takes the following params: <multisig address>, <start epoch>, <end epoch>
func (a *multiSig) MsigGetVested(ctx context.Context, addr address.Address, start types.TipSetKey, end types.TipSetKey) (types.BigInt, error) {
	startTS, err := a.state.ChainGetTipSet(ctx, start)
	if err != nil {
		return types.EmptyInt, fmt.Errorf("loading start tipset %s: %w", start, err)
	}

	endTS, err := a.state.ChainGetTipSet(ctx, end)
	if err != nil {
		return types.EmptyInt, fmt.Errorf("loading end tipset %s: %w", end, err)
	}

	if startTS.Height() > endTS.Height() {
		return types.EmptyInt, fmt.Errorf("start tipset %d is after end tipset %d", startTS.Height(), endTS.Height())
	} else if startTS.Height() == endTS.Height() {
		return big.Zero(), nil
	}

	// LoadActor(ctx, addr, endTs)
	act, err := a.state.GetParentStateRootActor(ctx, endTS, addr)
	if err != nil {
		return types.EmptyInt, fmt.Errorf("failed to load multisig actor at end epoch: %w", err)
	}

	msas, err := multisig.Load(a.Store(ctx), act)
	if err != nil {
		return types.EmptyInt, fmt.Errorf("failed to load multisig actor state: %w", err)
	}

	startLk, err := msas.LockedBalance(startTS.Height())
	if err != nil {
		return types.EmptyInt, fmt.Errorf("failed to compute locked balance at start height: %w", err)
	}

	endLk, err := msas.LockedBalance(endTS.Height())
	if err != nil {
		return types.EmptyInt, fmt.Errorf("failed to compute locked balance at end height: %w", err)
	}

	return types.BigSub(startLk, endLk), nil
}

func (a *multiSig) msigApproveOrCancelSimple(ctx context.Context, operation MsigProposeResponse, msig address.Address, txID uint64, src address.Address) (*types.MessagePrototype, error) {
	if msig == address.Undef {
		return nil, fmt.Errorf("must provide multisig address")
	}

	if src == address.Undef {
		return nil, fmt.Errorf("must provide source address")
	}

	mb, err := a.messageBuilder(ctx, src)
	if err != nil {
		return nil, err
	}

	var msg *types.Message
	switch operation {
	case MsigApprove:
		msg, err = mb.Approve(msig, txID, nil)
	case MsigCancel:
		msg, err = mb.Cancel(msig, txID, nil)
	default:
		return nil, fmt.Errorf("invalid operation for msigApproveOrCancel")
	}
	if err != nil {
		return nil, err
	}

	return &types.MessagePrototype{Message: *msg, ValidNonce: false}, nil
}

func (a *multiSig) msigApproveOrCancelTxnHash(ctx context.Context, operation MsigProposeResponse, msig address.Address, txID uint64, proposer address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.MessagePrototype, error) {
	if msig == address.Undef {
		return nil, fmt.Errorf("must provide multisig address")
	}

	if src == address.Undef {
		return nil, fmt.Errorf("must provide source address")
	}

	if proposer.Protocol() != address.ID {
		proposerID, err := a.state.StateLookupID(ctx, proposer, types.EmptyTSK)
		if err != nil {
			return nil, err
		}
		proposer = proposerID
	}

	p := multisig.ProposalHashData{
		Requester: proposer,
		To:        to,
		Value:     amt,
		Method:    abi.MethodNum(method),
		Params:    params,
	}

	mb, err := a.messageBuilder(ctx, src)
	if err != nil {
		return nil, err
	}

	var msg *types.Message
	switch operation {
	case MsigApprove:
		msg, err = mb.Approve(msig, txID, &p)
	case MsigCancel:
		msg, err = mb.Cancel(msig, txID, &p)
	default:
		return nil, fmt.Errorf("invalid operation for msigApproveOrCancel")
	}
	if err != nil {
		return nil, err
	}

	return &types.MessagePrototype{
		Message:    *msg,
		ValidNonce: false,
	}, nil
}

func (a *multiSig) MsigGetAvailableBalance(ctx context.Context, addr address.Address, tsk types.TipSetKey) (types.BigInt, error) {
	ts, err := a.state.ChainGetTipSet(ctx, tsk)
	if err != nil {
		return types.EmptyInt, fmt.Errorf("failed to load tipset: %w", err)
	}
	act, err := a.state.StateGetActor(ctx, addr, tsk)
	if err != nil {
		return types.EmptyInt, fmt.Errorf("failed to load multisig actor: %w", err)
	}
	msas, err := multisig.Load(a.Store(ctx), act)
	if err != nil {
		return types.EmptyInt, fmt.Errorf("failed to load multisig actor state: %w", err)
	}
	locked, err := msas.LockedBalance(ts.Height())
	if err != nil {
		return types.EmptyInt, fmt.Errorf("failed to compute locked multisig balance: %w", err)
	}
	return types.BigSub(act.Balance, locked), nil
}

func (a *multiSig) MsigGetVestingSchedule(ctx context.Context, addr address.Address, tsk types.TipSetKey) (types.MsigVesting, error) {
	act, err := a.state.StateGetActor(ctx, addr, tsk)
	if err != nil {
		return types.EmptyVesting, fmt.Errorf("failed to load multisig actor: %w", err)
	}

	msas, err := multisig.Load(a.Store(ctx), act)
	if err != nil {
		return types.EmptyVesting, fmt.Errorf("failed to load multisig actor state: %w", err)
	}

	ib, err := msas.InitialBalance()
	if err != nil {
		return types.EmptyVesting, fmt.Errorf("failed to load multisig initial balance: %w", err)
	}

	se, err := msas.StartEpoch()
	if err != nil {
		return types.EmptyVesting, fmt.Errorf("failed to load multisig start epoch: %w", err)
	}

	ud, err := msas.UnlockDuration()
	if err != nil {
		return types.EmptyVesting, fmt.Errorf("failed to load multisig unlock duration: %w", err)
	}

	return types.MsigVesting{
		InitialBalance: ib,
		StartEpoch:     se,
		UnlockDuration: ud,
	}, nil
}

func (a *multiSig) MsigGetPending(ctx context.Context, addr address.Address, tsk types.TipSetKey) ([]*types.MsigTransaction, error) {
	act, err := a.state.StateGetActor(ctx, addr, tsk)
	if err != nil {
		return nil, fmt.Errorf("failed to load multisig actor: %w", err)
	}
	msas, err := multisig.Load(a.Store(ctx), act)
	if err != nil {
		return nil, fmt.Errorf("failed to load multisig actor state: %w", err)
	}

	var out = []*types.MsigTransaction{}
	if err := msas.ForEachPendingTxn(func(id int64, txn multisig.Transaction) error {
		out = append(out, &types.MsigTransaction{
			ID:     id,
			To:     txn.To,
			Value:  txn.Value,
			Method: txn.Method,
			Params: txn.Params,

			Approved: txn.Approved,
		})
		return nil
	}); err != nil {
		return nil, err
	}

	return out, nil
}

func serializeAddParams(new address.Address, inc bool) ([]byte, error) {
	enc, actErr := actors.SerializeParams(&multisig2.AddSignerParams{
		Signer:   new,
		Increase: inc,
	})
	if actErr != nil {
		return nil, actErr
	}

	return enc, nil
}

func serializeSwapParams(old address.Address, new address.Address) ([]byte, error) {
	enc, actErr := actors.SerializeParams(&multisig2.SwapSignerParams{
		From: old,
		To:   new,
	})
	if actErr != nil {
		return nil, actErr
	}

	return enc, nil
}

func serializeRemoveParams(rem address.Address, dec bool) ([]byte, error) {
	enc, actErr := actors.SerializeParams(&multisig2.RemoveSignerParams{
		Signer:   rem,
		Decrease: dec,
	})
	if actErr != nil {
		return nil, actErr
	}

	return enc, nil
}
