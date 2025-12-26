package dto

import (
	"encoding/json"
	"net/netip"

	"github.com/docker/go-units"
)

type Rule struct {
	Prefix    netip.Prefix
	RateLimit int64
}

func (r Rule) MarshalJSON() ([]byte, error) {
	return json.Marshal(RuleJSON{
		Prefix:    r.Prefix.String(),
		RateLimit: units.HumanSize(float64(r.RateLimit)),
	})
}

type RuleJSON struct {
	Prefix    string `json:"prefix"`
	RateLimit string `json:"rate_limit"`
}
