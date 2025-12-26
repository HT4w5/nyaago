package analyzer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/netip"
	"time"
)

// Modify this after changing Record struct
const recEncodedSize = 136 // Observed value

type Record struct {
	Addr         netip.Addr
	Bucket       int64
	LastModified time.Time
}

func (r Record) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&r); err != nil {
		return nil, fmt.Errorf("failed to encode record: %w", err)
	}
	return buf.Bytes(), nil
}

func (r Record) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&r); err != nil {
		return fmt.Errorf("failed to decode record: %w", err)
	}
	return nil
}
