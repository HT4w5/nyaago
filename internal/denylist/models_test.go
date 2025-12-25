package denylist

import (
	"net/netip"
	"testing"
)

func TestEntryMarshalUnmarshalRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		prefix netip.Prefix
		blame  netip.Addr
	}{
		{
			name:   "IPv6 full",
			prefix: netip.MustParsePrefix("2001:db8::/32"),
			blame:  netip.MustParseAddr("2001:db8::1"),
		},
		{
			name:   "IPv6 single address",
			prefix: netip.MustParsePrefix("2001:db8::1/128"),
			blame:  netip.MustParseAddr("2001:db8::2"),
		},
		{
			name:   "IPv4 mapped to IPv6",
			prefix: netip.MustParsePrefix("192.168.1.0/24"),
			blame:  netip.MustParseAddr("10.0.0.1"),
		},
		{
			name:   "IPv4 single address",
			prefix: netip.MustParsePrefix("192.168.1.1/32"),
			blame:  netip.MustParseAddr("192.168.1.254"),
		},
		{
			name:   "IPv6 with zero bits",
			prefix: netip.MustParsePrefix("::/0"),
			blame:  netip.MustParseAddr("::1"),
		},
		{
			name:   "IPv6 with max bits",
			prefix: netip.MustParsePrefix("2001:db8::/128"),
			blame:  netip.MustParseAddr("2001:db8::"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &Entry{
				Prefix: tt.prefix,
				Blame:  tt.blame,
			}

			// Marshal the entry
			data := entry.Marshal()
			if len(data) != entSizeTotal {
				t.Errorf("Marshal() length = %d, want %d", len(data), entSizeTotal)
			}

			// Unmarshal into a new entry
			var newEntry Entry
			if err := newEntry.Unmarshal(data); err != nil {
				t.Errorf("Unmarshal() error = %v", err)
			}

			// Compare
			if newEntry.Prefix != entry.Prefix {
				t.Errorf("Unmarshal() prefix = %v, want %v", newEntry.Prefix, entry.Prefix)
			}
			if newEntry.Blame != entry.Blame {
				t.Errorf("Unmarshal() blame = %v, want %v", newEntry.Blame, entry.Blame)
			}
		})
	}
}

func TestEntryUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "nil data",
			data: nil,
		},
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "short data",
			data: make([]byte, entSizeTotal-1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var entry Entry
			if err := entry.Unmarshal(tt.data); err == nil {
				t.Errorf("Unmarshal() expected error, got nil")
			}
		})
	}
}

func TestEntryMarshalConsistency(t *testing.T) {
	// Test that marshaling the same entry multiple times produces the same output
	entry := &Entry{
		Prefix: netip.MustParsePrefix("2001:db8::/32"),
		Blame:  netip.MustParseAddr("2001:db8::1"),
	}

	data1 := entry.Marshal()
	data2 := entry.Marshal()

	if len(data1) != len(data2) {
		t.Fatalf("two marshals produced different lengths: %d vs %d", len(data1), len(data2))
	}

	for i := range data1 {
		if data1[i] != data2[i] {
			t.Fatalf("byte at index %d differs: %02x vs %02x", i, data1[i], data2[i])
		}
	}
}

func TestEntryUnmarshalThenMarshal(t *testing.T) {
	// Create a known good byte representation manually
	prefix := netip.MustParsePrefix("2001:db8::/32")
	blame := netip.MustParseAddr("2001:db8::1")

	entry := &Entry{Prefix: prefix, Blame: blame}
	data := entry.Marshal()

	var unmarshaled Entry
	if err := unmarshaled.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Remarshal the unmarshaled entry
	data2 := unmarshaled.Marshal()

	// Compare the two byte slices
	if len(data) != len(data2) {
		t.Fatalf("length mismatch: %d vs %d", len(data), len(data2))
	}
	for i := range data {
		if data[i] != data2[i] {
			t.Fatalf("byte at index %d differs: %02x vs %02x", i, data[i], data2[i])
		}
	}
}
