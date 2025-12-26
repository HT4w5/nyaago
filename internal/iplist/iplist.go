package iplist

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/allegro/bigcache/v3"
)

type IPList struct {
	cfg   *config.Config
	cache *bigcache.BigCache
}

func MakeIPList(cfg *config.Config) (*IPList, error) {
	l := &IPList{
		cfg: cfg,
	}

	var err error
	l.cache, err = bigcache.New(context.Background(), bigcache.DefaultConfig(cfg.IPList.EntryTTL.Duration))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return l, nil
}

func (l *IPList) ListRules() ([]dto.Rule, error) {
	it := l.cache.Iterator()
	rules := make([]dto.Rule, 0, l.cache.Len())
	prefixRateLimitMap := make(map[netip.Prefix]int64)
	for it.SetNext() {
		v, err := it.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get entry: %w", err)
		}
		var e IPEntry
		err = e.UnmarshalBinary(v.Value())
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal entry: %w", err)
		}
		if !e.Valid {
			continue
		}

		// Get prefix
		prefixLength := 32
		if e.Addr.Is4() {
			prefixLength = l.cfg.IPList.ExportPrefixLength.IPv4
		} else if e.Addr.Is6() {
			prefixLength = l.cfg.IPList.ExportPrefixLength.IPv6
		}
		prefix := netip.PrefixFrom(e.Addr, prefixLength).Masked()
		oldRateLimit, ok := prefixRateLimitMap[prefix]
		if !ok || e.RateLimit > oldRateLimit {
			prefixRateLimitMap[prefix] = e.RateLimit
		}
	}

	for k, v := range prefixRateLimitMap {
		rules = append(rules, dto.Rule{
			Prefix:    k,
			RateLimit: v,
		})
	}

	return rules, nil
}

func (l *IPList) Close() error {
	err := l.cache.Close()
	if err != nil {
		return fmt.Errorf("failed to close cache")
	}
	return nil
}
