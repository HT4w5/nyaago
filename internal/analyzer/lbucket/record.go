package lbucket

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/netip"
	"time"
)

type record struct {
	Addr         netip.Addr
	Bucket       int64
	LastModified time.Time
}

func (r *record) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encode record: %w", err)
	}
	return buf.Bytes(), nil
}

func (r *record) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(r); err != nil {
		return fmt.Errorf("failed to decode record: %w", err)
	}
	return nil
}

func (r record) DBKey() []byte {
	s := r.Addr.As16()
	return s[:]
}
