package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Log      LogConfig      `json:"log"`
	DB       DBConfig       `json:"db"`
	Analyzer AnalyzerConfig `json:"analyzer"`
	Ingress  IngressConfig  `json:"ingress"`
	Egress   EgressConfig   `json:"egress"`
	API      APIConfig      `json:"api"`
}

func Load(path string) (*Config, error) {
	cfgBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := getDefault()

	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	err = cfg.verify()
	if err != nil {
		return nil, fmt.Errorf("failed to verify config: %w", err)
	}
	return &cfg, nil
}

func getDefault() Config {
	var cfg Config

	cfg.Analyzer.BanPrefixLength.IPv4 = 24
	cfg.Analyzer.BanPrefixLength.IPv6 = 64

	cfg.API.ListenAddr = "0.0.0.0:80"

	return cfg
}

func (cfg Config) verify() error {
	if cfg.Analyzer.BanPrefixLength.IPv4 < 0 || cfg.Analyzer.BanPrefixLength.IPv4 >= 32 {
		return fmt.Errorf("invalid analyzer.ban_prefix_length.ipv4. Must be in range of [0, 32]")
	}
	if cfg.Analyzer.BanPrefixLength.IPv6 < 0 || cfg.Analyzer.BanPrefixLength.IPv6 >= 128 {
		return fmt.Errorf("invalid analyzer.ban_prefix_length.ipv6. Must be in range of [0, 128]")
	}
	return nil
}
