package dto

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
