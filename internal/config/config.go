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

	// Ingress
	cfg.Ingress.Syslog.Transport = "udp"
	cfg.Ingress.Syslog.ListenAddr = "0.0.0.0:514"

	// RuleList
	cfg.RuleList.EntryTTL = Duration(30 * time.Minute)
	cfg.RuleList.ExportPrefixLength.IPv4 = 24
	cfg.RuleList.ExportPrefixLength.IPv6 = 64

	// API
	cfg.API.ListenAddr = "0.0.0.0:8580"

	return cfg
}

func (cfg Config) verify() error {
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
