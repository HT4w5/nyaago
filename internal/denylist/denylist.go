package denylist

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/allegro/bigcache/v3"
	"go4.org/netipx"
)

const (
	slogModuleName = "denylist"
	slogGroupName  = "denylist"
)

type DenyList struct {
	cfg    *config.Config
	cache  *bigcache.BigCache
	logger *slog.Logger
}

func MakeDenyList(cfg *config.Config) (*DenyList, error) {
	l := &DenyList{
		cfg: cfg,
	}

	logger, err := logging.GetLogger(&cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}
	l.logger = logger.With(logging.SlogKeyModule, slogModuleName).WithGroup(slogGroupName)

	l.cache, err = bigcache.New(context.Background(), bigcache.DefaultConfig(cfg.DenyList.EntryTTL.Duration))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return l, nil
}

func (l *DenyList) PutEntry(p netip.Prefix, b netip.Addr) error {
	err := l.putEntry(Entry{
		Prefix: p,
		Blame:  b,
	})
	if err != nil {
		l.logger.Error("failed to put entry", logging.SlogKeyError, err)
		return fmt.Errorf("failed to put entry: %w", err)
	}
	return nil
}

func (l *DenyList) GetEntry(b netip.Addr) (Entry, error) {
	e, err := l.getEntry(b)
	if err != nil {
		l.logger.Error("failed to get entry", logging.SlogKeyError, err)
		return Entry{}, fmt.Errorf("failed to get entry: %w", err)
	}

	return e, nil
}

func (l *DenyList) GetIPSet() (*netipx.IPSet, error) {
	var b netipx.IPSetBuilder
	it := l.cache.Iterator()
	for it.SetNext() {
		v, err := it.Value()
		if err != nil {
			l.logger.Error("failed to get entry", logging.SlogKeyError, err)
			continue
		}
		var entry Entry
		err = entry.Unmarshal(v.Value())
		if err != nil {
			l.logger.Error("failed to unmarshal entry", logging.SlogKeyError, err)
			continue
		}
		b.AddPrefix(entry.Prefix)
	}

	return b.IPSet()
}

func (l *DenyList) ListRules() ([]dto.Rule, error) {
	it := l.cache.Iterator()
	rules := make([]dto.Rule, 0, l.cache.Len())
	for it.SetNext() {
		v, err := it.Value()
		if err != nil {
			l.logger.Error("failed to get entry", logging.SlogKeyError, err)
			return nil, fmt.Errorf("failed to get entry: %w", err)
		}
		var entry Entry
		err = entry.Unmarshal(v.Value())
		if err != nil {
			l.logger.Error("failed to unmarshal entry", logging.SlogKeyError, err)
			continue
		}
		r := dto.Rule{
			Prefix:    entry.Prefix,
			Addr:      entry.Blame,
			ExpiresOn: time.Unix(int64(v.Timestamp()), 0).Add(l.cfg.DenyList.EntryTTL.Duration),
		}
		rules = append(rules, r)
	}

	return rules, nil
}

func (l *DenyList) Close() {
	err := l.cache.Close()
	if err != nil {
		l.logger.Error("failed to close cache", logging.SlogKeyError, err)
	}
}
