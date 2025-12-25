package dto

import (
	"fmt"
	"net/netip"
	"time"
)

type Rule struct {
	Prefix    netip.Prefix
	Addr      netip.Addr
	ExpiresOn time.Time
}

func (r Rule) JSON() RuleJSON {
	return RuleJSON{
		Prefix:    r.Prefix.String(),
		Addr:      r.Addr.String(),
		ExpiresOn: r.ExpiresOn.Unix(),
	}
}

type RuleJSON struct {
	Prefix    string `json:"prefix"`
	Addr      string `json:"addr"`
	ExpiresOn int64  `json:"expires_on"` // Unix timestamp in seconds
}

// Omit prefix
func (r RuleJSON) ToObject() (Rule, error) {
	var rule Rule
	var err error
	rule.Addr, err = netip.ParseAddr(r.Addr)
	if err != nil {
		return Rule{}, fmt.Errorf("failed to parse addr")
	}

	rule.ExpiresOn = time.Unix(r.ExpiresOn, 0)

	return rule, nil
}
