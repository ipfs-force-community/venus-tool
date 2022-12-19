package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/ipfs-force-community/venus-tool/route"
)

type Client struct {
	*resty.Client
}

func New(url string) (*Client, error) {
	client := resty.New().
		SetHostURL(url).
		SetHeader("Accept", "application/json")
	_, err := client.R().Get("/version")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", url, err)
	}
	return &Client{Client: client}, nil
}

func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	errResp := &route.ErrorResp{}
	_, err := c.R().SetContext(ctx).SetBody(body).SetResult(result).SetError(errResp).Post(path)
	if err != nil {
		return err
	}
	if errResp.Err != "" {
		return errResp
	}
	return nil
}
