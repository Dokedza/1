package cfg

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port         int    `json:"port"`
	WorkersCount int    `json:"workers_count"`
	StorePath    string `json:"store_path"`
	Timeout      int    `json:"timeout"`
}

func Load() *Config {
	cfg := &Config{
		Port:         8080,
		WorkersCount: 5,
		StorePath:    "storage.json",
		Timeout:      1000,
	}

	file, err := os.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(file, cfg)
	}

	return cfg
}
