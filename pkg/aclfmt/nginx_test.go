package aclfmt

import (
	"bytes"
	"fmt"
	"net/netip"
	"testing"

	"github.com/HT4w5/nyaago/pkg/utils"
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
			ipset: utils.MakeIPSet(
				[]netip.Prefix{
					netip.MustParsePrefix("192.168.1.0/24"),
					netip.MustParsePrefix("10.0.0.0/8"),
				},
				[]netip.Prefix{},
			),
			expectError: false,
		},
		{
			name: "info with new line",
			info: "nyaago\nnewline",
			ipset: utils.MakeIPSet(
				[]netip.Prefix{
					netip.MustParsePrefix("192.168.1.0/24"),
				},
				[]netip.Prefix{},
			),
			expectError: true,
		},
		{
			name: "empty info",
			info: "",
			ipset: utils.MakeIPSet(
				[]netip.Prefix{
					netip.MustParsePrefix("192.168.1.0/24"),
				},
				[]netip.Prefix{},
			),
			expectError: true,
		},
		{
			name: "empty ipset",
			info: "nyaago",
			ipset: utils.MakeIPSet(
				[]netip.Prefix{},
				[]netip.Prefix{},
			),
			expectError: false,
		},
		{
			name: "single ip",
			info: "nyaago",
			ipset: utils.MakeIPSet(
				[]netip.Prefix{
					netip.MustParsePrefix("192.168.1.1/32"),
				},
				[]netip.Prefix{},
			),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := makeNginxFormatter(tt.info)
			err := f.Marshal(tt.ipset, &buf)
			if err != nil && !tt.expectError {
				t.Errorf("unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("expected error, got none")
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
