package config

import (
	"bytes"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Path        string `toml:"-"`
	Server      *ServerConfig
	NodeAPI     APIInfo
	MessagerAPI APIInfo
	MarketAPI   APIInfo
	WalletAPI   APIInfo
	AuthAPI     APIInfo
	DamoclesAPI APIInfo
}

type APIInfo struct {
	Addr  string
	Token string
}

type ServerConfig struct {
	ListenAddr string
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
		Server: &ServerConfig{
			ListenAddr: "127.0.0.1:12580",
		},
	}
}
