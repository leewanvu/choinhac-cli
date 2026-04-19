package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MusicDir string `yaml:"music_dir"`
	DBPath   string `yaml:"db_path"`
	Port     int    `yaml:"port"`
	BindAddr string `yaml:"bind_addr"`
}

func Default() Config {
	home, _ := os.UserHomeDir()
	return Config{
		MusicDir: filepath.Join(home, "Music"),
		DBPath:   filepath.Join(home, ".config", "musiccli", "library.db"),
		Port:     8080,
		BindAddr: "127.0.0.1",
	}
}

func Load() (Config, error) {
	cfg := Default()
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "musiccli", "config.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}
