package project

import (
	"os"
	"path/filepath"
)

func FindConfig(dir string) (string, *Config, error) {
	current := dir
	for {
		path := filepath.Join(current, "portmap.toml")
		if _, err := os.Stat(path); err == nil {
			cfg, err := LoadConfig(path)
			if err != nil {
				return "", nil, err
			}
			return path, cfg, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			// reached the filesystem root without finding a config
			break
		}
		current = parent
	}
	return "", nil, nil
}

func InferName(dir string) string {
	_, cfg, _ := FindConfig(dir)
	if cfg != nil && cfg.Project != "" {
		return cfg.Project
	}
	return filepath.Base(dir)
}
