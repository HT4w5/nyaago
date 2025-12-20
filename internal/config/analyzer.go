package config

type AnalyzerConfig struct {
	TTL                Duration        `json:"ttl"`
	UpdateInterval     Duration        `json:"update_interval"`
	SendRatioThreshold float64         `json:"send_ratio_threshold"`
	Include            []RequestFilter `json:"include"`
	Exclude            []RequestFilter `json:"exclude"`
	BanPrefixLength    struct {
		IPv4 int `json:"ipv4"`
		IPv6 int `json:"ipv6"`
	} `json:"ban_prefix_length"`
}
