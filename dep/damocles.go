package dep

import (
	"github.com/ipfs-force-community/damocles/damocles-manager/core"
)

type IDamocles interface {
	core.SealerCliAPI
}
