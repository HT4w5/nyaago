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
	Prefix    netip.Prefix
	Banned    bool
	RateLimit int64
	Blame     string
	ExpiresAt time.Time
}

func (e *Rule) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(e); err != nil {
		return nil, fmt.Errorf("failed to encode entry: %w", err)
	}
	return buf.Bytes(), nil
}

func (e *Rule) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(e); err != nil {
		return fmt.Errorf("failed to decode entry: %w", err)
	}
	return nil
}

// Create fixed-length []byte key for a Rule
func (e Rule) DBKey() []byte {
	b := make([]byte, 17)
	prefix := e.Prefix.Masked()
	addr := prefix.Addr().As16()
	copy(b[0:16], addr[:])
	b[16] = uint8(prefix.Bits())
	return b
}

func (r Rule) MarshalJSON() ([]byte, error) {
	return json.Marshal(RuleJSON{
		Prefix:    r.Prefix.String(),
		Banned:    r.Banned,
		RateLimit: units.HumanSize(float64(r.RateLimit)),
		Blame:     r.Blame,
		ExpiresAt: r.ExpiresAt.Format(time.RFC3339),
	})
}

type RuleJSON struct {
	Prefix    string `json:"prefix"`
	Banned    bool   `json:"banned"`
	RateLimit string `json:"rate_limit"`
	Blame     string `json:"blame"`
	ExpiresAt string `json:"expires_at"`
}
