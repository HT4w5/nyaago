package dto

import (
	"net/netip"
	"time"
)

/*
One request parsed from logs
*/
type Request struct {
	Client   netip.Addr // IPv4 or IPv6
	URL      string     // E.g. "/foo/bar"
	BodySent int64
	Time     time.Time
}
