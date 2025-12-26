package iplist

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/netip"
	"time"
)

type IPEntry struct {
	Valid     bool
	Addr      netip.Addr
	RateLimit int64
	ExpiresOn time.Time
}

func (e IPEntry) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&e); err != nil {
		return nil, fmt.Errorf("failed to encode entry: %w", err)
	}
	return buf.Bytes(), nil
}

func (e IPEntry) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&e); err != nil {
		return fmt.Errorf("failed to decode entry: %w", err)
	}
	return nil
}
