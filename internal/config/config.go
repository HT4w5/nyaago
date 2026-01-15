package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type ConfigObject interface {
	setDefault()     // Sets all fields to default value
	validate() error // Verify fields
}

type Config struct {
	Log      LogConfig      `json:"log"`
	DB       DBConfig       `json:"db"`
	Router   RouterConfig   `json:"router"`
	RuleList RuleListConfig `json:"ip_list"`
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

	// DB
	cfg.DB.Dir = "./nyaago.db"

	// Router
	cfg.Router.RecordTTL.Duration = 2 * time.Hour
	// 10MB/s
	cfg.Router.LeakRate = 10000000
	// 6GB
	cfg.Router.Capacity = 6000000000
	cfg.Router.Cache.Shards = 1024
	cfg.Router.Cache.CleanInterval.Duration = 5 * time.Minute
	cfg.Router.Cache.RPS = 10
	// 1GB
	cfg.Router.Cache.MaxSize = 1000000000
	cfg.Router.FilterMode = "none"

	// Ingress
	cfg.Ingress.Syslog.Transport = "udp"
	cfg.Ingress.Syslog.ListenAddr = "0.0.0.0:514"

	// RuleList
	cfg.RuleList.EntryTTL.Duration = 30 * time.Minute
	cfg.RuleList.ExportPrefixLength.IPv4 = 24
	cfg.RuleList.ExportPrefixLength.IPv6 = 64

	// API
	cfg.API.ListenAddr = "0.0.0.0:8580"

	return cfg
}

func (cfg Config) verify() error {
	// RuleList
	if cfg.RuleList.ExportPrefixLength.IPv4 < 0 || cfg.RuleList.ExportPrefixLength.IPv4 >= 32 {
		return fmt.Errorf("invalid ip_list.export_prefix_length.ipv4. Must be in range of [0, 32]")
	}
	if cfg.RuleList.ExportPrefixLength.IPv6 < 0 || cfg.RuleList.ExportPrefixLength.IPv6 >= 128 {
		return fmt.Errorf("invalid ip_list.export_prefix_length.ipv6. Must be in range of [0, 128]")
	}

	// Router
	if !inValidList(cfg.Router.FilterMode, routerFilterModes) {
		return fmt.Errorf("invalid router.filter_mode. Must be one of %v", routerFilterModes)
	}

	if cfg.Router.Cache.MaxSize < 1000000 && cfg.Router.Cache.MaxSize == 0 {
		return fmt.Errorf("invalid router.cache.max_size. Must be larger than 1MB.")
	}
	return nil
}

func inValidList(s string, l []string) bool {
	for _, v := range l {
		if v == s {
			return true
		}
	}
	return false
}
