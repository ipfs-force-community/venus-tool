package dep

import (
	nodeV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/api/market"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	"go.uber.org/fx"
)

type ServiceParams struct {
	fx.In
	Messager messager.IMessager
	Market   market.IMarket
	Wallet   IWallet
	Node     nodeV1.FullNode
	Auth     IAuth
}

type IWallet interface {
	wallet.ICommon
	wallet.IWallet
}
