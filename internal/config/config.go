package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Log      LogConfig      `json:"log"`
	Analyzer AnalyzerConfig `json:"analyzer"`
	DenyList DenyListConfig `json:"deny_list"`
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

	// Analyzer
	cfg.Analyzer.BanPrefixLength.IPv4 = 24
	cfg.Analyzer.BanPrefixLength.IPv6 = 64
	cfg.Analyzer.RecordTTL.Duration = 24 * time.Hour
	// 100Mbps
	cfg.Analyzer.LeakRate = 12500000
	// 500MB
	cfg.Analyzer.Capacity = 500000000
	cfg.Analyzer.Cache.Shards = 1024
	cfg.Analyzer.Cache.CleanInterval.Duration = 5 * time.Minute
	cfg.Analyzer.Cache.RPS = 10
	// 1GB
	cfg.Analyzer.Cache.MaxSize = 1000000000

	// API
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
