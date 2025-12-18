package dto

import (
	"net/netip"
	"time"
)

type Rule struct {
	Prefix    netip.Prefix
	Addr      netip.Addr
	URL       string
	ExpiresOn time.Time
}
