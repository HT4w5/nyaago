package config

import (
	"encoding/binary"
	"encoding/json"
	"hash/fnv"
)

type AnalyzerConfig interface {
	Hash() uint32 // Generate unique db key prefix for every unique config
}

type AnalyzerInstanceConfig struct {
	Tag    string
	Type   string
	Config AnalyzerConfig
}

type AnalyzerInstanceConfigJSON struct {
	Tag    string          `json:"tag"`
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

func (c *AnalyzerInstanceConfig) UnmarshalJSON(data []byte) error {
	var cfgJSON AnalyzerInstanceConfigJSON
	err := json.Unmarshal(data, &cfgJSON)
	if err != nil {
		return err
	}
	c.Tag = cfgJSON.Tag
	c.Type = cfgJSON.Type

	switch cfgJSON.Type {
	case "lbucket":
		var lc LeakyBucketConfig
		err := json.Unmarshal(cfgJSON.Config, &lc)
		if err != nil {
			return err
		}
		c.Config = lc
	default:
	}
	return nil
}

func (c AnalyzerInstanceConfig) Hash() uint32 {
	h := fnv.New32a()
	binary.Write(h, binary.BigEndian, c.Tag)
	binary.Write(h, binary.BigEndian, c.Type)
	return hashCombine(h.Sum32(), c.Config.Hash())
}

type LeakyBucketConfig struct {
	LeakRate ByteSize     `json:"leak_rate"`
	Capacity ByteSize     `json:"capacity"`
	MinRate  ByteSize     `json:"min_rate"`
	Export   ExportConfig `json:"export"`
}

func (c LeakyBucketConfig) Hash() uint32 {
	h := fnv.New32a()
	binary.Write(h, binary.BigEndian, c.LeakRate)
	binary.Write(h, binary.BigEndian, c.Capacity)
	binary.Write(h, binary.BigEndian, c.MinRate)
	return hashCombine(h.Sum32(), c.Export.Hash())
}

type ExportConfig struct {
	PrefixLength struct {
		IPv4 int `json:"ipv4"`
		IPv6 int `json:"ipv6"`
	} `json:"prefix_length"`
	TTL Duration `json:"ttl"`
}

func (c ExportConfig) Hash() uint32 {
	h := fnv.New32a()
	binary.Write(h, binary.BigEndian, c.PrefixLength.IPv4)
	binary.Write(h, binary.BigEndian, c.PrefixLength.IPv6)
	binary.Write(h, binary.BigEndian, c.TTL)
	return h.Sum32()
}

func hashCombine(h1, h2 uint32) uint32 {
	return h1 ^ (h2 + 0x9e3779b9 + (h1 << 6) + (h1 >> 2))
}
