package project

import (
	"os"

	"github.com/BurntSushi/toml"
)

type PortEntry struct {
	Port        int    `toml:"port"`
	Service     string `toml:"service"`
	Description string `toml:"description"`
}

type Config struct {
	Project string      `toml:"project"`
	Ports   []PortEntry `toml:"ports"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func WriteConfig(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}
