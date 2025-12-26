package iplist

import (
	"context"
	"fmt"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/allegro/bigcache/v3"
	"go4.org/netipx"
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
	l.cache, err = bigcache.New(context.Background(), bigcache.DefaultConfig(cfg.IPList.RuleTTL.Duration))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return l, nil
}

func (l *IPList) GetIPSet() (*netipx.IPSet, error) {
	var b netipx.IPSetBuilder
	it := l.cache.Iterator()
	for it.SetNext() {
		v, err := it.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get entry: %w", err)
		}
		var rule dto.Rule
		err = rule.UnmarshalBinary(v.Value())
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal entry: %w", err)
		}
		if rule.Valid {
			b.AddPrefix(rule.Prefix)
		}
	}

	return b.IPSet()
}

func (l *IPList) ListRules() ([]dto.Rule, error) {
	it := l.cache.Iterator()
	rules := make([]dto.Rule, 0, l.cache.Len())
	for it.SetNext() {
		v, err := it.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get entry: %w", err)
		}
		var r dto.Rule
		err = r.UnmarshalBinary(v.Value())
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal entry: %w", err)
		}
		if !r.Valid {
			continue
		}
		rules = append(rules, r)
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
