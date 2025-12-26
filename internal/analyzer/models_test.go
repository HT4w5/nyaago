package analyzer

import (
	"math/rand"
	"net/netip"
	"testing"
	"time"
)

func TestRecordMarshalUnmarshal(t *testing.T) {
	testCases := []struct {
		name      string
		record    Record
		expectErr bool
	}{
		{
			"Valid IPv4 address",
			Record{Addr: netip.MustParseAddr("192.168.1.1"), Bucket: 123, LastModified: time.Unix(1672531200, 0)},
			false,
		},
		{
			"Valid IPv6 address",
			Record{Addr: netip.MustParseAddr("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), Bucket: -456, LastModified: time.Now()},
			false,
		},
		{
			"Edge case: minimum bucket value",
			Record{Addr: netip.MustParseAddr("127.0.0.1"), Bucket: -9223372036854775808, LastModified: time.Now()},
			false,
		},
		{
			"Edge case: maximum bucket value",
			Record{Addr: netip.MustParseAddr("::1"), Bucket: 9223372036854775807, LastModified: time.Now()},
			false,
		},
		{
			"Edge case: minimum time value",
			Record{Addr: netip.MustParseAddr("0.0.0.0"), Bucket: 100, LastModified: time.Unix(0, 0)},
			false,
		},
		{
			"Edge case: maximum time value",
			Record{Addr: netip.MustParseAddr("255.255.255.255"), Bucket: 200, LastModified: time.Unix(0, 9223372036854775807)},
			false,
		},
		{
			"Invalid data length",
			Record{},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectErr {
				err := tc.record.Unmarshal(make([]byte, 0))
				if err == nil {
					t.Errorf("Expected an error, got none")
				}
			} else {
				marshaled, _ := tc.record.Marshal()
				unmarshaled := &Record{}
				err := unmarshaled.Unmarshal(marshaled)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if unmarshaled.Addr != tc.record.Addr {
					t.Errorf("Addr mismatch: expected %v, got %v", tc.record.Addr, unmarshaled.Addr)
				}

				if unmarshaled.Bucket != tc.record.Bucket {
					t.Errorf("Bucket mismatch: expected %v, got %v", tc.record.Bucket, unmarshaled.Bucket)
				}

				if unmarshaled.LastModified.UnixNano() != tc.record.LastModified.UnixNano() {
					t.Errorf("LastModified mismatch: expected %v, got %v", tc.record.LastModified, unmarshaled.LastModified)
				}
			}
		})
	}
}

func TestEncodedSize(t *testing.T) {
	const numRecords = 10000
	var totalSize int64

	for i := 0; i < numRecords; i++ {
		// Generate random IP (v4 or v6)
		var ip netip.Addr
		if rand.Intn(2) == 0 {
			// IPv4
			ip = netip.AddrFrom4([4]byte{
				byte(rand.Intn(256)),
				byte(rand.Intn(256)),
				byte(rand.Intn(256)),
				byte(rand.Intn(256)),
			})
		} else {
			// IPv6
			ip = netip.AddrFrom16([16]byte{
				byte(rand.Intn(256)), byte(rand.Intn(256)),
				byte(rand.Intn(256)), byte(rand.Intn(256)),
				byte(rand.Intn(256)), byte(rand.Intn(256)),
				byte(rand.Intn(256)), byte(rand.Intn(256)),
				byte(rand.Intn(256)), byte(rand.Intn(256)),
				byte(rand.Intn(256)), byte(rand.Intn(256)),
				byte(rand.Intn(256)), byte(rand.Intn(256)),
				byte(rand.Intn(256)), byte(rand.Intn(256)),
			})
		}

		// Generate random bucket value
		bucket := rand.Int63() - (1 << 62) // Full int64 range

		// Generate random timestamp within last 10 years
		now := time.Now()
		randomTime := now.Add(-time.Duration(rand.Int63n(10*365*24*60*60)) * time.Second)

		record := Record{
			Addr:         ip,
			Bucket:       bucket,
			LastModified: randomTime,
		}

		data, err := record.Marshal()
		if err != nil {
			t.Errorf("Failed to marshal record: %v", err)
			return
		}
		totalSize += int64(len(data))
	}

	averageSize := float64(totalSize) / float64(numRecords)
	t.Logf("Generated %d random records", numRecords)
	t.Logf("Total encoded size: %d bytes", totalSize)
	t.Logf("Average encoded size: %.2f bytes", averageSize)
}
