package dto

import (
	"net/netip"
	"time"
)

type Rule struct {
	Prefix    netip.Prefix
	Addr      netip.Addr
	URL       string
	ExpiresOn time.Time
}

type RuleJSON struct {
	Prefix    string `json:"prefix"`
	Addr      string `json:"addr"`
	URL       string `json:"url"`
	ExpiresOn int64  `json:"expires_on"` // Unix timestamp in seconds
}

func (r Rule) JSON() RuleJSON {
	return RuleJSON{
		Prefix:    r.Prefix.String(),
		Addr:      r.Addr.String(),
		URL:       r.URL,
		ExpiresOn: r.ExpiresOn.Unix(),
	}
}
