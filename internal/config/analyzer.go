package config

type AnalyzerConfig struct {
	RecordTTL Duration `json:"record_ttl"`
	// Bucket leak rate (bytes per second)
	LeakRate         ByteSize        `json:"leak_rate"`
	Capacity         ByteSize        `json:"capacity"`
	Include          []RequestFilter `json:"include"`
	Exclude          []RequestFilter `json:"exclude"`
	DenyPrefixLength struct {
		IPv4 int `json:"ipv4"`
		IPv6 int `json:"ipv6"`
	} `json:"deny_prefix_length"`
	Cache CacheConfig `json:"cache"`
}

type CacheConfig struct {
	Shards        int      `json:"shards"`
	CleanInterval Duration `json:"clean_interval"`
	RPS           int      `json:"rps"`
	MaxSize       ByteSize `json:"max_size"`
}
