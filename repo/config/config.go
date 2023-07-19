package config

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/BurntSushi/toml"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

type Config struct {
	Path        string `toml:"-"`
	Server      ServerConfig
	NodeAPI     APIInfo
	MessagerAPI APIInfo
	MarketAPI   APIInfo
	WalletAPI   APIInfo
	AuthAPI     APIInfo
	DamoclesAPI APIInfo
	MinerAPI    APIInfo
}

type ServerConfig struct {
	ListenAddr string
	BoardPath  string
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	cfg.Path = path
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = toml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Save() error {
	b := bytes.Buffer{}
	err := toml.NewEncoder(&b).Encode(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.Path, b.Bytes(), 0644)
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			ListenAddr: "127.0.0.1:8090",
			BoardPath:  "./dashboard/build",
		},
	}
}

type APIInfo struct {
	Addr  string
	Token string
}

func (a APIInfo) DialArgs(version string) (string, error) {
	ma, err := multiaddr.NewMultiaddr(a.Addr)
	if err == nil {
		_, addr, err := manet.DialArgs(ma)
		if err != nil {
			return "", err
		}

		return "ws://" + addr + "/rpc/" + version, nil
	}

	_, err = url.Parse(a.Addr)
	if err != nil {
		return "", err
	}
	return a.Addr + "/rpc/" + version, nil
}

func (a APIInfo) AuthHeader() http.Header {
	if len(a.Token) != 0 {
		headers := http.Header{}
		headers.Add("Authorization", "Bearer "+string(a.Token))
		return headers
	}
	return nil
}
