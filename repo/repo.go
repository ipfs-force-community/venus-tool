package repo

import (
	"os"
	"path/filepath"

	"github.com/ipfs-force-community/venus-tool/repo/config"
)

const ConfigPath = "config"

type Repo struct {
	Path string
}

func NewRepo(path string) *Repo {
	repo := &Repo{
		Path: path,
	}
	return repo
}

func (r *Repo) Exists() bool {
	// check if repo path exists
	if _, err := os.Stat(r.Path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (r *Repo) Init(cfg *config.Config) error {
	// create repo path
	if err := os.MkdirAll(r.Path, os.ModePerm); err != nil {
		return err
	}
	// create config file
	cfgPath := filepath.Join(r.Path, ConfigPath+".toml")
	cfg.Path = cfgPath
	return cfg.Save()
}

func (r *Repo) GetPath() string {
	return r.Path
}

func (r *Repo) GetConfig() (*config.Config, error) {
	cfgPath := filepath.Join(r.Path, ConfigPath+".toml")
	return config.LoadConfig(cfgPath)
}
