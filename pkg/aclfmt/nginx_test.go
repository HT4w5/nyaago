package aclfmt

import (
	"bytes"
	"fmt"
	"net/netip"
	"testing"

	"go4.org/netipx"
)

func TestMakeNginxFormatter(t *testing.T) {
	tests := []struct {
		name     string
		info     string
		expected string
	}{
		{
			name:     "happy path",
			info:     "nyaago",
			expected: "nyaago",
		},
		{
			name:     "info with new line",
			info:     "nyaago\nnewline",
			expected: "",
		},
		{
			name:     "empty info",
			info:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := makeNginxFormatter(tt.info)
			if f.info != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, f.info)
			}
		})
	}
}

func TestNginxFormatterMarshal(t *testing.T) {
	tests := []struct {
		name        string
		info        string
		ipset       *netipx.IPSet
		expectError bool
	}{
		{
			name: "happy path",
			info: "nyaago",
			ipset: makeIPSet([]string{
				"192.168.1.0/24",
				"10.0.0.0/8",
			}),
		},
		{
			name: "info with new line",
			info: "nyaago\nnewline",
			ipset: makeIPSet([]string{
				"192.168.1.0/24",
			}),
		},
		{
			name: "empty info",
			info: "",
			ipset: makeIPSet([]string{
				"192.168.1.0/24",
			}),
		},
		{
			name:  "empty ipset",
			info:  "nyaago",
			ipset: makeIPSet([]string{}),
		},
		{
			name: "single ip",
			info: "nyaago",
			ipset: makeIPSet([]string{
				"192.168.1.1/32",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := makeNginxFormatter(tt.info)
			err := f.Marshal(tt.ipset, &buf)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError {
				fmt.Println(buf.String())
			}
		})
	}
}

func TestNginxFormatterInfo(t *testing.T) {
	tests := []struct {
		name     string
		info     string
		expected string
	}{
		{
			name:     "happy path",
			info:     "nyaago",
			expected: "nyaago",
		},
		{
			name:     "info with new line",
			info:     "nyaago\nnewline",
			expected: "",
		},
		{
			name:     "empty info",
			info:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := makeNginxFormatter(tt.info)
			result := f.Info()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func makeIPSet(prefixes []string) *netipx.IPSet {
	var b netipx.IPSetBuilder
	for _, v := range prefixes {
		p, err := netip.ParsePrefix(v)
		if err != nil {
			return nil
		}

		b.AddPrefix(p)
	}

	set, _ := b.IPSet()
	return set
}
