package denylist

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/allegro/bigcache/v3"
	"go4.org/netipx"
)

type DenyList struct {
	cfg   *config.Config
	cache *bigcache.BigCache
}

func MakeDenyList(cfg *config.Config) (*DenyList, error) {
	l := &DenyList{
		cfg: cfg,
	}

	var err error
	l.cache, err = bigcache.New(context.Background(), bigcache.DefaultConfig(cfg.DenyList.RuleTTL.Duration))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return l, nil
}

func (l *DenyList) PutRule(r dto.Rule) error {
	if !r.Valid {
		return nil
	}
	err := l.putRule(r)
	if err != nil {
		return fmt.Errorf("failed to put entry: %w", err)
	}
	return nil
}

func (l *DenyList) GetRule(b netip.Addr) (dto.Rule, error) {
	e, err := l.getRule(b)
	if err != nil {
		return dto.Rule{}, fmt.Errorf("failed to get entry: %w", err)
	}
	if !e.Valid {
		return dto.Rule{}, nil
	}

	return e, nil
}

func (l *DenyList) DelRule(b netip.Addr) error {
	err := l.putRule(dto.Rule{
		Valid: false,
	})
	if err != nil {
		return fmt.Errorf("failed to put entry: %w", err)
	}
	return nil
}

func (l *DenyList) GetIPSet() (*netipx.IPSet, error) {
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

func (l *DenyList) ListRules() ([]dto.Rule, error) {
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

func (l *DenyList) Close() error {
	err := l.cache.Close()
	if err != nil {
		return fmt.Errorf("failed to close cache")
	}
	return nil
}
