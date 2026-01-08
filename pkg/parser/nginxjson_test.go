package parser

import (
	"testing"
	"time"
)

func TestParseNginxTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTime time.Time
		wantErr  bool
	}{
		{
			name:     "valid time with 3 decimal places",
			input:    "1734345934.123",
			wantTime: time.Unix(1734345934, 123000000),
			wantErr:  false,
		},
		{
			name:     "valid time with 1 decimal place",
			input:    "1734345934.1",
			wantTime: time.Unix(1734345934, 100000000),
			wantErr:  false,
		},
		{
			name:     "valid time with 6 decimal places",
			input:    "1734345934.123456",
			wantTime: time.Unix(1734345934, 123456000),
			wantErr:  false,
		},
		{
			name:     "valid time with 9 decimal places (nanoseconds)",
			input:    "1734345934.123456789",
			wantTime: time.Unix(1734345934, 123456789),
			wantErr:  false,
		},
		{
			name:     "time with no decimal part",
			input:    "1734345934",
			wantTime: time.Time{},
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			wantTime: time.Time{},
			wantErr:  true,
		},
		{
			name:     "invalid format - no dot",
			input:    "invalid",
			wantTime: time.Time{},
			wantErr:  true,
		},
		{
			name:     "time with only decimal part",
			input:    ".123",
			wantTime: time.Unix(0, 123000000),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNginxTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNginxTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.wantTime) {
				t.Errorf("parseNginxTime() = %v, want %v", got, tt.wantTime)
			}
		})
	}
}

func TestParseNginxDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantDur time.Duration
		wantErr bool
	}{
		{
			name:    "valid duration with 3 decimal places",
			input:   "1.123",
			wantDur: time.Duration(1123000000), // 1.123 seconds in nanoseconds
			wantErr: false,
		},
		{
			name:    "valid duration with 1 decimal place",
			input:   "0.5",
			wantDur: time.Duration(500000000), // 0.5 seconds in nanoseconds
			wantErr: false,
		},
		{
			name:    "valid duration with 6 decimal places",
			input:   "2.123456",
			wantDur: time.Duration(2123456000), // 2.123456 seconds in nanoseconds
			wantErr: false,
		},
		{
			name:    "valid duration with 9 decimal places",
			input:   "0.123456789",
			wantDur: time.Duration(123456789), // 0.123456789 seconds in nanoseconds
			wantErr: false,
		},
		{
			name:    "duration with no decimal part",
			input:   "5",
			wantDur: 0,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantDur: 0,
			wantErr: true,
		},
		{
			name:    "invalid format - no dot",
			input:   "invalid",
			wantDur: 0,
			wantErr: true,
		},
		{
			name:    "duration with only decimal part",
			input:   ".123",
			wantDur: time.Duration(123000000), // 0.123 seconds in nanoseconds
			wantErr: true,
		},
		{
			name:    "zero duration",
			input:   "0.0",
			wantDur: time.Duration(0),
			wantErr: false,
		},
		{
			name:    "large duration",
			input:   "86400.123",                   // 1 day + 123ms
			wantDur: time.Duration(86400123000000), // 86400.123 seconds in nanoseconds
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNginxDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNginxDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantDur {
				t.Errorf("parseNginxDuration() = %v, want %v", got, tt.wantDur)
			}
		})
	}
}
