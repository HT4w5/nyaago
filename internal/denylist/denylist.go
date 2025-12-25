package denylist

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/logging"
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

func (l *DenyList) Put(p netip.Prefix, b netip.Addr) {
	err := l.putEntry(Entry{
		Prefix: p,
		Blame:  b,
	})
	if err != nil {
		l.logger.Error("failed to put entry", logging.SlogKeyError, err)
	}
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
