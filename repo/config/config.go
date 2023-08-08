package config

import (
	"bytes"
	"net/http"
	"net/url"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

type Config struct {
	Path   string `toml:"-"`
	Server ServerConfig

	ChainService *APIInfo

	NodeAPI     *APIInfo
	MessagerAPI *APIInfo
	MarketAPI   *APIInfo
	AuthAPI     *APIInfo
	MinerAPI    *APIInfo

	WalletAPI   APIInfo
	DamoclesAPI APIInfo
}

func mergeAPIInfo(prior, alternative *APIInfo) *APIInfo {
	if prior == nil {
		return alternative
	}
	if alternative == nil {
		return prior
	}
	if prior.Addr == "" {
		prior.Addr = alternative.Addr
	}
	if prior.Token == "" {
		prior.Token = alternative.Token
	}
	return prior
}

func (c Config) GetNodeAPI() APIInfo {
	p := mergeAPIInfo(c.NodeAPI, c.ChainService)
	if p == nil {
		return APIInfo{}
	}
	return *p
}

func (c Config) GetMessagerAPI() APIInfo {
	p := mergeAPIInfo(c.MessagerAPI, c.ChainService)
	if p == nil {
		return APIInfo{}
	}
	return *p
}

func (c Config) GetMarketAPI() APIInfo {
	p := mergeAPIInfo(c.MarketAPI, c.ChainService)
	if p == nil {
		return APIInfo{}
	}
	return *p
}

func (c Config) GetAuthAPI() APIInfo {
	p := mergeAPIInfo(c.AuthAPI, c.ChainService)
	if p == nil {
		return APIInfo{}
	}
	return *p
}

func (c Config) GetMinerAPI() APIInfo {
	p := mergeAPIInfo(c.MinerAPI, c.ChainService)
	if p == nil {
		return APIInfo{}
	}
	return *p
}

type ServerConfig struct {
	ListenAddr string
	BoardPath  string
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	cfg.Path = path
	data, err := os.ReadFile(path)
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
	return os.WriteFile(c.Path, b.Bytes(), 0644)
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
