package dto

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	"github.com/docker/go-units"
)

type Rule struct {
	Valid     bool
	Blame     netip.Addr
	Prefix    netip.Prefix
	RateLimit int64
	ExpiresOn time.Time
}

func (r *Rule) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encode entry: %w", err)
	}
	return buf.Bytes(), nil
}

func (r *Rule) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(r); err != nil {
		return fmt.Errorf("failed to decode entry: %w", err)
	}
	return nil
}

func (r Rule) MarshalJSON() ([]byte, error) {
	return json.Marshal(RuleJSON{
		Blame:     r.Blame.String(),
		Prefix:    r.Prefix.String(),
		RateLimit: units.HumanSize(float64(r.RateLimit)),
		ExpiresOn: r.ExpiresOn.UTC().Format(time.RFC3339),
	})
}

type RuleJSON struct {
	Blame     string `json:"blame"`
	Prefix    string `json:"prefix"`
	RateLimit string `json:"rate_limit"`
	ExpiresOn string `json:"expires_on"` // RFC3399
}
