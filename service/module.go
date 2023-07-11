package service

import (
	"github.com/ipfs-force-community/venus-tool/dep"
	"github.com/ipfs-force-community/venus-tool/pkg/multisig"
)

func NewService(params dep.ServiceParams) (*ServiceImpl, error) {
	return &ServiceImpl{
		Messager: params.Messager,
		Market:   params.Market,
		Node:     params.Node,
		Wallet:   params.Wallet,
		Auth:     params.Auth,
		Damocles: params.Damocles,
		Miner:    params.Miner,

		Multisig: multisig.NewMultiSig(params.Node),
	}, nil
}
