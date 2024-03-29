package config

import (
	"bytes"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/filecoin-project/venus/venus-shared/api"
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
	return api.DialArgs(a.Addr, version)
}

func (a APIInfo) AuthHeader() http.Header {
	if len(a.Token) != 0 {
		headers := http.Header{}
		headers.Add("Authorization", "Bearer "+string(a.Token))
		return headers
	}
	return nil
}
