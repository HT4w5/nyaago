package config

type IPListConfig struct {
	EntryTTL           Duration `json:"entry_ttl"`
	ExportPrefixLength struct {
		IPv4 int `json:"ipv4"`
		IPv6 int `json:"ipv6"`
	} `json:"export_prefix_length"`
}
