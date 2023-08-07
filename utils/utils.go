package utils

import (
	"context"
	"fmt"
	"net/url"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/utils"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multiaddr"
	maNet "github.com/multiformats/go-multiaddr/net"
)

type networkNameKeeper struct {
	NetworkName types.NetworkName
}

func (nk *networkNameKeeper) StateNetworkName(ctx context.Context) (types.NetworkName, error) {
	return nk.NetworkName, nil
}

func LoadBuiltinActors(ctx context.Context, networkName types.NetworkName) error {
	nk := &networkNameKeeper{NetworkName: networkName}
	return utils.LoadBuiltinActors(ctx, nk)
}

func GetMethodMeta(actorCode cid.Cid, method abi.MethodNum) (utils.MethodMeta, error) {
	methodMeta, found := utils.MethodsMap[actorCode][method]
	if !found {
		return utils.MethodMeta{}, fmt.Errorf("method %d not found on actor %s", method, actorCode)
	}
	return methodMeta, nil
}

// ParseAddr parse a multi addr to a traditional url ( with http scheme as default)
func ParseAddr(addr string) (string, error) {
	ret := addr
	ma, err := multiaddr.NewMultiaddr(addr)
	if err == nil {
		_, addr, err := maNet.DialArgs(ma)
		if err != nil {
			return "", fmt.Errorf("parser libp2p url fail %w", err)
		}

		ret = "http://" + addr

		_, err = ma.ValueForProtocol(multiaddr.P_WSS)
		if err == nil {
			ret = "wss://" + addr
		} else if err != multiaddr.ErrProtocolNotFound {
			return "", err
		}

		_, err = ma.ValueForProtocol(multiaddr.P_HTTPS)
		if err == nil {
			ret = "https://" + addr
		} else if err != multiaddr.ErrProtocolNotFound {
			return "", err
		}

		_, err = ma.ValueForProtocol(multiaddr.P_WS)
		if err == nil {
			ret = "ws://" + addr
		} else if err != multiaddr.ErrProtocolNotFound {
			return "", err
		}

		_, err = ma.ValueForProtocol(multiaddr.P_HTTP)
		if err == nil {
			ret = "http://" + addr
		} else if err != multiaddr.ErrProtocolNotFound {
			return "", err
		}
	}

	_, err = url.Parse(ret)
	if err != nil {
		return "", fmt.Errorf("parser address fail %w", err)
	}

	return ret, nil
}
