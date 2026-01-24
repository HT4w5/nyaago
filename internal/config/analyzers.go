package config

// Config for request analyzer
type AnaylzerConfig struct {
	LeakyBucket      LeakyBucketConfig      `json:"leaky_bucket"`
	FileSendRatio    FileSendRatioConfig    `json:"file_send_ratio"`
	RequestFrequency RequestFrequencyConfig `json:"request_frequecy"`
}

// Config for leaky bucket analyzer
type LeakyBucketConfig struct {
	Enabled   bool     `json:"enabled"`
	LeakRate  ByteSize `json:"leak_rate"`  // Rate of bucket leak per second for a single client
	Capacity  ByteSize `json:"capacity"`   // Capacity of bucket. Amount of data accumulated before a bucket leaks
	BucketTTL Duration `json:"bucket_ttl"` // Record's time to live
	Export    struct {
		ExportCommonConfig
		MinRate ByteSize `json:"min_rate"` // Minimum rate limit applyed to a client (to avoid connection timeout)
	}
}

type FileSendRatioConfig struct {
	Enabled   bool     `json:"enabled"`
	UnitTime  Duration `json:"unit_time"`  // Analysis duration of a single record
	RecordTTL Duration `json:"record_ttl"` // Record's time to live
	PathMap   []struct {
		UrlPrefix string `json:"url_prefix"`
		DirPrefix string `json:"dir_prefix"`
	} `json:"path_map"` // Path mapping from URL to filesystem for file size indexing
	SizeInfoTTL Duration `json:"size_info_ttl"` // Size info's time to live
	Export      struct {
		ExportCommonConfig
		RatioThreshold float64 `json:"ratio_threshold"` // Minimum send ratio for a client to be exported. Any client with a living ratio record larger than this will be banned
	}
}

type RequestFrequencyConfig struct {
	Enabled      bool     `json:"enabled"`
	UnitTime     Duration `json:"unit_time"`     // Analysis duration of a single record
	RecordTTL    Duration `json:"record_ttl"`    // Record's time to live
	RPSThreshold float64  `json:"rps_threshold"` // Max request per second allowed for a client. Any client with a living rps record larger than this will be banned
	Export       struct {
		ExportCommonConfig
	}
}

// Common configs for rule export
type ExportCommonConfig struct {
	PrefixLength struct {
		IPv4 int `json:"ipv4"` // Affected range for IPv4
		IPv6 int `json:"ipv6"` // Affected range for IPv6
	} `json:"prefix_length"`
	TTL Duration `json:"ttl"` // Exported rule's time to live
}
