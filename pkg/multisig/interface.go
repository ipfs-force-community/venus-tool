package multisig

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"

	"github.com/filecoin-project/venus/venus-shared/types"
)

type IMultiSig interface {
	MsigCreate(ctx context.Context, req uint64, addrs []address.Address, duration abi.ChainEpoch, val types.BigInt, src address.Address, gp types.BigInt) (*types.MessagePrototype, error)
	StateMsigInfo(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*types.MsigInfo, error)
	MsigPropose(ctx context.Context, msig address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.MessagePrototype, error)
	MsigAddPropose(ctx context.Context, msig address.Address, src address.Address, newAdd address.Address, inc bool) (*types.MessagePrototype, error)
	MsigAddApprove(ctx context.Context, msig address.Address, src address.Address, txID uint64, proposer address.Address, newAdd address.Address, inc bool) (*types.MessagePrototype, error)
	MsigAddCancel(ctx context.Context, msig address.Address, src address.Address, txID uint64, newAdd address.Address, inc bool) (*types.MessagePrototype, error)
	// MsigCancel cancels a previously-proposed multisig message
	// It takes the following params: <multisig address>, <proposed transaction ID>, <recipient address>, <value to transfer>,
	// <sender address of the cancel msg>, <method to call in the proposed message>, <params to include in the proposed message>
	MsigCancelTxnHash(context.Context, address.Address, uint64, address.Address, types.BigInt, address.Address, uint64, []byte) (*types.MessagePrototype, error)
	MsigSwapPropose(ctx context.Context, msig address.Address, src address.Address, oldAdd address.Address, newAdd address.Address) (*types.MessagePrototype, error)
	MsigSwapApprove(ctx context.Context, msig address.Address, src address.Address, txID uint64, proposer address.Address, oldAdd address.Address, newAdd address.Address) (*types.MessagePrototype, error)
	MsigSwapCancel(ctx context.Context, msig address.Address, src address.Address, txID uint64, oldAdd address.Address, newAdd address.Address) (*types.MessagePrototype, error)
	MsigApprove(ctx context.Context, msig address.Address, txID uint64, src address.Address) (*types.MessagePrototype, error)
	MsigApproveTxnHash(ctx context.Context, msig address.Address, txID uint64, proposer address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.MessagePrototype, error)
	MsigCancel(ctx context.Context, msig address.Address, txID uint64, src address.Address) (*types.MessagePrototype, error)
	MsigRemoveSigner(ctx context.Context, msig address.Address, proposer address.Address, toRemove address.Address, decrease bool) (*types.MessagePrototype, error)

	// MsigGetVested returns the amount of FIL that vested in a multisig in a certain period.
	// It takes the following params: <multisig address>, <start epoch>, <end epoch>
	MsigGetVested(context.Context, address.Address, types.TipSetKey, types.TipSetKey) (types.BigInt, error)
	// MsigGetAvailableBalance returns the portion of a multisig's balance that can be withdrawn or spent
	MsigGetAvailableBalance(context.Context, address.Address, types.TipSetKey) (types.BigInt, error)
	// MsigGetVestingSchedule returns the vesting details of a given multisig.
	MsigGetVestingSchedule(context.Context, address.Address, types.TipSetKey) (types.MsigVesting, error)

	//MsigGetPending returns pending transactions for the given multisig
	//wallet. Once pending transactions are fully approved, they will no longer
	//appear here.
	MsigGetPending(context.Context, address.Address, types.TipSetKey) ([]*types.MsigTransaction, error)
}
