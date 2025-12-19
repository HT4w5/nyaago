package config

import (
	"net/netip"
)

type AnalyzerConfig struct {
	TTL                Duration `json:"ttl"`
	UpdateInterval     Duration `json:"update_interval"`
	SendRatioThreshold float64  `json:"send_ratio_threshold"`
	Include            []Filter `json:"include"`
	Exclude            []Filter `json:"exclude"`
	BanPrefixLength    struct {
		IPv4 int `json:"ipv4"`
		IPv6 int `json:"ipv6"`
	} `json:"ban_prefix_length"`
}

type Filter struct {
	Prefix   netip.Prefix `json:"prefix"`
	URLRegex RegexWrapper `json:"url_regex"`
}
