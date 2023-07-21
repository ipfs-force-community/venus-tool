package dep

import (
	"context"
	"fmt"

	"github.com/ipfs-force-community/sophon-auth/jwtclient"
	"github.com/ipfs-force-community/venus-tool/repo/config"
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
	jwt, err := jwtclient.NewAuthClient(cfg.AuthAPI.Addr, cfg.AuthAPI.Token)
	if err != nil {
		return nil, err
	}

	playLoad, err := jwt.Verify(ctx, cfg.AuthAPI.Token)
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
