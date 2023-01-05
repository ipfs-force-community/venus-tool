package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-resty/resty/v2"
	"github.com/ipfs-force-community/venus-tool/route"
)

type Client struct {
	*resty.Client
	apiVersion string
}

func New(url string) (*Client, error) {
	client := resty.New().
		SetHostURL(url).
		SetHeader("Accept", "application/json")

	_, err := client.R().Get("/version")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", url, err)
	}

	client.SetAllowGetMethodPayload(true)

	return &Client{Client: client}, nil
}

func (c *Client) SetVersion(version string) {
	c.apiVersion = version
}

func (c *Client) Do(ctx context.Context, method, path string, params, result interface{}) error {
	path = c.apiVersion + path

	errResp := &route.ErrorResp{}
	req := c.R().SetContext(ctx).SetError(errResp)
	if params != nil {
		switch method {
		case http.MethodGet:
			if m, ok := params.(map[string]string); ok {
				req.SetQueryParams(m)
			} else if m, err := toMap(params, false); err == nil {
				req.SetQueryParams(m)
			} else {
				req.SetBody(params)
			}
		case http.MethodDelete:
			if m, ok := params.(map[string]string); ok {
				req.SetQueryParams(m)
			} else if m, err := toMap(params, true); err == nil {
				req.SetQueryParams(m)
			} else {
				panic("parse params failed")
			}
		default:
			req.SetBody(params)
		}
	}
	if result != nil {
		req = req.SetResult(result)
	}

	resp, err := req.Execute(method, path)
	if err != nil {
		return err
	}
	if errResp.Err != "" {
		return errResp
	}
	if resp.IsError() {
		return fmt.Errorf("http error: %s", resp.Status())
	}
	return nil
}

// toMap converts a simple struct to a map[string]string
func toMap(params interface{}, allowStruct bool) (map[string]string, error) {
	ret := make(map[string]string)

	rtype := reflect.TypeOf(params)
	rvalue := reflect.ValueOf(params)

	if rtype.Kind() == reflect.Ptr {
		rtype = rtype.Elem()
		rvalue = rvalue.Elem()
	}

	if rtype.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", rtype.Kind())
	}

	for i := 0; i < rtype.NumField(); i++ {
		field := rtype.Field(i)
		value := rvalue.Field(i)

		if field.Type.Kind() == reflect.Ptr {
			if value.IsNil() {
				continue
			}
			value = value.Elem()
		}

		if !value.CanSet() {
			return nil, fmt.Errorf("unexported field %s", field.Name)
		}

		// convert to string
		var str string
		switch field.Type.Kind() {
		case reflect.String:
			str = value.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			str = fmt.Sprintf("%d", value.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			str = fmt.Sprintf("%d", value.Uint())
		case reflect.Bool:
			str = fmt.Sprintf("%t", value.Bool())
		case reflect.Struct:
			if !allowStruct {
				return nil, fmt.Errorf("structs not allowed")
			}
			b, err := json.Marshal(value.Interface())
			if err != nil {
				return nil, err
			}
			str = string(b)
		default:
			return nil, fmt.Errorf("unsupported type %s", field.Type.Kind())
		}
		ret[field.Name] = str
	}
	return ret, nil
}
