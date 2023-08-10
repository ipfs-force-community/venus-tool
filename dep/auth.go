package dep

import (
	"context"
	"fmt"

	"github.com/ipfs-force-community/sophon-auth/jwtclient"
	"github.com/ipfs-force-community/venus-tool/repo/config"
	"github.com/ipfs-force-community/venus-tool/utils"
)

type IAuth interface {
	jwtclient.IAuthClient
	GetUserName(ctx context.Context) (string, error)
}

type auth struct {
	jwtclient.IAuthClient
	name string
}

func (a *auth) GetUserName(ctx context.Context) (string, error) {
	return a.name, nil
}

func NewAuth(ctx context.Context, cfg *config.Config) (IAuth, error) {
	// todo parse url from multiaddr
	authToken := cfg.GetAuthAPI().Token
	authAddr, err := utils.ParseAddr(cfg.GetAuthAPI().Addr)
	if err != nil {
		return nil, err
	}
	jwt, err := jwtclient.NewAuthClient(authAddr, authToken)
	if err != nil {
		return nil, err
	}

	playLoad, err := jwt.Verify(ctx, authToken)
	if err != nil {
		return nil, err
	}

	userName := playLoad.Name
	if userName == "" {
		return nil, fmt.Errorf("user from token is empty")
	}

	return &auth{
		jwt,
		userName,
	}, nil
}
