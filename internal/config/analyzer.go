package config

type AnalyzerConfig struct {
	TTL       Duration `json:"ttl"`
	RecordTTL Duration `json:"record_ttl"`
	// Bucket leak rate (bytes per second)
	LeakRate           ByteSize        `json:"leak_rate"`
	Capacity           ByteSize        `json:"capacity"`
	UpdateInterval     Duration        `json:"update_interval"`
	SendRatioThreshold float64         `json:"send_ratio_threshold"`
	Include            []RequestFilter `json:"include"`
	Exclude            []RequestFilter `json:"exclude"`
	BanPrefixLength    struct {
		IPv4 int `json:"ipv4"`
		IPv6 int `json:"ipv6"`
	} `json:"ban_prefix_length"`
	Cache CacheConfig `json:"cache"`
}

type CacheConfig struct {
	Shards        int      `json:"shards"`
	CleanInterval Duration `json:"clean_interval"`
	MaxSize       ByteSize `json:"max_size"`
}
