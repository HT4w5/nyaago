package config

type Config struct {
	Log  LogConfig  `json:"log"`
	DB   DBConfig   `json:"db"`
	Pool PoolConfig `json:"pool"`
	Tail TailConfig `json:"tail"`
	Fmt  FmtConfig  `json:"fmt"`
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

func Load(path string) *Config {
	return nil
}
