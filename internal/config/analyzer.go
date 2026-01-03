package config

const (
	AnalyzerFilterModeIncludeOnly    = "include-only"
	AnalyzerFilterModeExcludeOnly    = "exclude-only"
	AnalyzerFilterModeIncludeExclude = "include-exclude"
	AnalyzerFilterModeExcludeInclude = "exclude-include"
)

var analyzerFilterModes = []string{
	AnalyzerFilterModeIncludeOnly,
	AnalyzerFilterModeExcludeOnly,
	AnalyzerFilterModeIncludeExclude,
	AnalyzerFilterModeExcludeInclude,
}

type AnalyzerConfig struct {
	RecordTTL Duration `json:"record_ttl"`
	// Bucket leak rate (bytes per second)
	LeakRate   ByteSize        `json:"leak_rate"`
	Capacity   ByteSize        `json:"capacity"`
	FilterMode string          `json:"filter_mode"`
	Include    []RequestFilter `json:"include"`
	Exclude    []RequestFilter `json:"exclude"`
	Cache      CacheConfig     `json:"cache"`
}

type CacheConfig struct {
	Shards        int      `json:"shards"`
	CleanInterval Duration `json:"clean_interval"`
	RPS           int      `json:"rps"`
	MaxSize       ByteSize `json:"max_size"`
}
