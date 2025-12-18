package db

import (
	"net/netip"
	"time"
)

type Client struct {
	Addr      netip.Addr
	TotalSent int64
	CreatedOn time.Time
	ExpiresOn time.Time
}

type Resource struct {
	URL       string
	Size      int64
	ExpiresOn time.Time
}

/*
Request represents a unique pair of client address and resource URL.
*/
type Request struct {
	Addr       netip.Addr // Client addr
	URL        string     //
	TotalSent  int64      // Total traffic sent to client since tracked
	SendRatio  float64    // TotalSent / <size of resource> / 1d
	Occurrence int
	CreatedOn  time.Time
	ExpiresOn  time.Time
}

type Rule struct {
	Prefix    netip.Prefix
	Addr      netip.Addr
	URL       string
	ExpiresOn time.Time
}
