package service

import (
	"github.com/filecoin-project/go-address"
	nodeV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/api/market"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	"github.com/ipfs-force-community/venus-tool/pkg/multisig"
	"go.uber.org/fx"
)

type ServiceParams struct {
	fx.In
	Messager messager.IMessager
	Market   market.IMarket
	Wallet   IWallet
	Node     nodeV1.FullNode
}

func (params ServiceParams) NewService(wallets, miners []address.Address) (*ServiceImpl, error) {
	return &ServiceImpl{
		Messager: params.Messager,
		Market:   params.Market,
		Node:     params.Node,
		Wallet:   params.Wallet,

		Multisig: multisig.NewMultiSig(params.Node),

		Wallets: wallets,
		Miners:  miners,
	}, nil
}

type IWallet interface {
	wallet.ICommon
	wallet.IWallet
}
