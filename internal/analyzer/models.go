package analyzer

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"time"
)

const (
	recOffAddr   = 0
	recOffBucket = 16
	recOffTime   = 24
	recSizeTotal = 32
)

type Record struct {
	Addr         netip.Addr
	Bucket       int64
	LastModified time.Time
}

/*
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
+                                                               +
|                                                               |
+                      IP Address (128 bits)                    +
|                (IPv4-mapped or Native IPv6)                   |
+                                                               +
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
+                        Bucket (64 bits)                       +
|                    (Signed 64-bit Integer)                    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
+                   Last Modified (64 bits)                     +
|                 (Unix Nanoseconds since Epoch)                |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

Total Length: 32 Octets (256 bits)
*/

func (r *Record) Marshal() []byte {
	buf := make([]byte, recSizeTotal)

	addr16 := r.Addr.As16()
	copy(buf[recOffAddr:recOffBucket], addr16[:])

	binary.BigEndian.PutUint64(buf[recOffBucket:recOffTime], uint64(r.Bucket))

	binary.BigEndian.PutUint64(buf[recOffTime:recSizeTotal], uint64(r.LastModified.UnixNano()))

	return buf
}

func (r *Record) Unmarshal(data []byte) error {
	if len(data) < recSizeTotal {
		return fmt.Errorf("data too short")
	}

	var addrBytes [recOffBucket]byte
	copy(addrBytes[:], data[recOffAddr:recOffBucket])
	r.Addr = netip.AddrFrom16(addrBytes).Unmap()

	r.Bucket = int64(binary.BigEndian.Uint64(data[recOffBucket:recOffTime]))

	nanos := int64(binary.BigEndian.Uint64(data[recOffTime:recSizeTotal]))
	r.LastModified = time.Unix(0, nanos)

	return nil
}
