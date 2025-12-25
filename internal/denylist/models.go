package denylist

import (
	"fmt"
	"net/netip"
)

const (
	// Offsets for the Entry binary format
	entOffPrefixAddr = 0
	entOffPrefixBits = 16
	entOffBlameAddr  = 17
	entSizeTotal     = 33
)

type Entry struct {
	Prefix netip.Prefix
	Blame  netip.Addr
}

/*
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
+                                                               +
|                                                               |
+                    Prefix IP (128 bits)                       +
|                                                               |
+                                                               +
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| Prefix Length |                                               |
+-+-+-+-+-+-+-+-+                                               +
|                                                               |
+                    Blame IP (128 bits)                        +
|                                                               |
+                                                               +
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

Total Length: 33 Octets
*/

// Marshal converts the Entry into a fixed-size byte slice.
func (e *Entry) Marshal() []byte {
	buf := make([]byte, entSizeTotal)

	pAddr16 := e.Prefix.Addr().As16()
	copy(buf[entOffPrefixAddr:entOffPrefixBits], pAddr16[:])

	buf[entOffPrefixBits] = uint8(e.Prefix.Bits())

	bAddr16 := e.Blame.As16()
	copy(buf[entOffBlameAddr:entSizeTotal], bAddr16[:])

	return buf
}

// Unmarshal populates the Entry from a byte slice.
func (e *Entry) Unmarshal(data []byte) error {
	if len(data) < entSizeTotal {
		return fmt.Errorf("data too short: expected %d, got %d", entSizeTotal, len(data))
	}

	var pAddrBytes [16]byte
	copy(pAddrBytes[:], data[entOffPrefixAddr:entOffPrefixBits])
	pAddr := netip.AddrFrom16(pAddrBytes).Unmap()
	bits := int(data[entOffPrefixBits])
	e.Prefix = netip.PrefixFrom(pAddr, bits)

	var bAddrBytes [16]byte
	copy(bAddrBytes[:], data[entOffBlameAddr:entSizeTotal])
	e.Blame = netip.AddrFrom16(bAddrBytes).Unmap()

	return nil
}
