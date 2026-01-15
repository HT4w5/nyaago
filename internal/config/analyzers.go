package config

type CacheConfig struct {
	Shards        int      `json:"shards"`
	CleanInterval Duration `json:"clean_interval"`
	RPS           int      `json:"rps"`
	MaxSize       ByteSize `json:"max_size"`
}

type LeakyBucketConfig struct {
	LeakRate     ByteSize   `json:"leak_rate"`
	Capacity     ByteSize   `json:"capacity"`
	MinRate      ByteSize   `json:"min_rate"`
	RuleSettings RuleConfig `json:"rule_settings"`
}

type RuleConfig struct {
	PrefixLength PrefixLengthConfig `json:"prefix_length"`
	TTL          Duration           `json:"ttl"`
}

type PrefixLengthConfig struct {
	IPv4 int `json:"ipv4"`
	IPv6 int `json:"ipv6"`
}
