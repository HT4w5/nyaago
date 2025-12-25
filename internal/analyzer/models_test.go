package analyzer

import (
	"fmt"
	"net/netip"
	"testing"
	"time"
)

func TestRecordMarshalUnmarshal(t *testing.T) {
	testCases := []struct {
		name        string
		record      Record
		expectedErr error
	}{
		{
			"Valid IPv4 address",
			Record{Addr: netip.MustParseAddr("192.168.1.1"), Bucket: 123, LastModified: time.Unix(1672531200, 0)},
			nil,
		},
		{
			"Valid IPv6 address",
			Record{Addr: netip.MustParseAddr("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), Bucket: -456, LastModified: time.Now()},
			nil,
		},
		{
			"Edge case: minimum bucket value",
			Record{Addr: netip.MustParseAddr("127.0.0.1"), Bucket: -9223372036854775808, LastModified: time.Now()},
			nil,
		},
		{
			"Edge case: maximum bucket value",
			Record{Addr: netip.MustParseAddr("::1"), Bucket: 9223372036854775807, LastModified: time.Now()},
			nil,
		},
		{
			"Edge case: minimum time value",
			Record{Addr: netip.MustParseAddr("0.0.0.0"), Bucket: 100, LastModified: time.Unix(0, 0)},
			nil,
		},
		{
			"Edge case: maximum time value",
			Record{Addr: netip.MustParseAddr("255.255.255.255"), Bucket: 200, LastModified: time.Unix(0, 9223372036854775807)},
			nil,
		},
		{
			"Invalid data length",
			Record{},
			fmt.Errorf("data too short"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedErr != nil {
				err := tc.record.Unmarshal(make([]byte, recOffTime))
				if err == nil {
					t.Errorf("Expected an error, got none: %v", tc.expectedErr)
				}
				if err.Error() != tc.expectedErr.Error() {
					t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
				}
			} else {
				marshaled := tc.record.Marshal()
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
