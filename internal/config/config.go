package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Log      LogConfig      `json:"log"`
	DB       DBConfig       `json:"db"`
	Pool     PoolConfig     `json:"pool"`
	Cron     CronConfig     `json:"cron"`
	Tail     TailConfig     `json:"tail"`
	Fmt      FmtConfig      `json:"fmt"`
	API      APIConfig      `json:"api"`
	PostExec PostExecConfig `json:"post_exec"`
}

type DBConfig struct {
	Type   string `json:"type"`
	Access string `json:"access"`
}

type LogConfig struct {
	Access   string `json:"access"`
	LogLevel string `json:"log_level"`
	Json     bool   `json:"json"`
}

type PoolConfig struct {
	ClientConfig       PoolObjectConfig  `json:"client"`
	ResourceConfig     PoolObjectConfig  `json:"resource"`
	RequestConfig      PoolRequestConfig `json:"request"`
	RuleConfig         PoolObjectConfig  `json:"rule"`
	SendRatioThreshold float64           `json:"send_ratio_threshold"`
	BanPrefixLength    struct {
		IPv4 int `json:"ipv4"`
		Ipv6 int `json:"ipv6"`
	} `json:"ban_prefix_length"`
}

type CronConfig struct {
	Interval Duration `json:"interval"`
}

type PoolObjectConfig struct {
	TTL       Duration `json:"ttl"`
	Whitelist []string `json:"whitelist"`
}

type PoolRequestConfig struct {
	MaturationThreshold int      `json:"maturation"`
	TTL                 Duration `json:"ttl"`
	Whitelist           []struct {
		Prefix string `json:"prefix"`
		URL    string `json:"url"`
	} `json:"whitelist"`
}

type TailConfig struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Poll bool   `json:"poll"`
}

type FmtConfig struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type APIConfig struct {
	Addr string `json:"addr"`
	Port int    `json:"port"`
}

type PostExecConfig struct {
	Cmd  string   `json:"cmd"`
	Cwd  string   `json:"cwd"`
	Args []string `json:"args"`
}

func Load(path string) (*Config, error) {
	cfgBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}
