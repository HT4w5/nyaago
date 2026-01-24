package fsr

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"net/netip"
	"time"

	"github.com/HT4w5/nyaago/internal/dbkey"
)

const (
	currentRecords dbkey.Prefix = iota
	historicRecords
	fileSizeRecords
	ipRecords
)

type ipRecord struct {
	Addr netip.Addr
}

func (r *ipRecord) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encode record: %w", err)
	}
	return buf.Bytes(), nil
}

func (r *ipRecord) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(r); err != nil {
		return fmt.Errorf("failed to decode record: %w", err)
	}
	return nil
}

func (r ipRecord) DBKey() []byte {
	addrBytes := r.Addr.As16()
	return addrBytes[:]
}

type currentRecord struct {
	Addr netip.Addr
	Path string
	Sent int64
}

func (r *currentRecord) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encode record: %w", err)
	}
	return buf.Bytes(), nil
}

func (r *currentRecord) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(r); err != nil {
		return fmt.Errorf("failed to decode record: %w", err)
	}
	return nil
}

func (r currentRecord) DBKey() []byte {
	addrBytes := r.Addr.As16()
	res := make([]byte, 0, 16+len(r.Path))
	res = append(res, addrBytes[:]...)
	res = append(res, r.Path...)
	return res
}

type historicRecord struct {
	Addr     netip.Addr
	Path     string
	Ratio    float64
	Time     time.Time
	Duration time.Duration
}

func (r *historicRecord) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encode record: %w", err)
	}
	return buf.Bytes(), nil
}

func (r *historicRecord) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(r); err != nil {
		return fmt.Errorf("failed to decode record: %w", err)
	}
	return nil
}

func (r historicRecord) DBKey() []byte {
	addrBytes := r.Addr.As16()
	res := make([]byte, 0, 24)
	res = append(res, addrBytes[:]...)
	binary.BigEndian.AppendUint64(res, uint64(r.Time.Unix()))
	return res
}

type fileSizeRecord struct {
	Path string
	Size int64
}

func (r *fileSizeRecord) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encode record: %w", err)
	}
	return buf.Bytes(), nil
}

func (r *fileSizeRecord) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(r); err != nil {
		return fmt.Errorf("failed to decode record: %w", err)
	}
	return nil
}

func (r fileSizeRecord) DBKey() []byte {
	return []byte(r.Path)
}
