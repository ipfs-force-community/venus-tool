package dep

import (
	"fmt"

	nodeV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	market "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
)

var log = logging.Logger("dep")

var ErrEmptyAddr = fmt.Errorf("empty api addr")

type ServiceParams struct {
	fx.In
	Messager messager.IMessager
	Market   market.IMarket
	Wallet   IWallet
	Node     nodeV1.FullNode
	Auth     IAuth
	Damocles *Damocles
	Miner    Miner
}

type IWallet interface {
	wallet.ICommon
	wallet.IWallet
}
