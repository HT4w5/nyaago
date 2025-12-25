package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/denylist"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/allegro/bigcache/v3"
)

const (
	slogModuleName = "analyzer"
	slogGroupName  = "analyzer"
)

type Analyzer struct {
	cfg      *config.Config
	cache    *bigcache.BigCache
	logger   *slog.Logger
	denylist *denylist.DenyList
}

func MakeAnalyzer(cfg *config.Config, denylist *denylist.DenyList) (*Analyzer, error) {
	a := &Analyzer{
		cfg:      cfg,
		denylist: denylist,
	}

	logger, err := logging.GetLogger(&cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}
	a.logger = logger.With(logging.SlogKeyModule, slogModuleName).WithGroup(slogGroupName)

	cacheConfig := bigcache.Config{
		Shards:             a.cfg.Analyzer.Cache.Shards,
		LifeWindow:         a.cfg.Analyzer.RecordTTL.Duration,
		CleanWindow:        a.cfg.Analyzer.Cache.CleanInterval.Duration,
		MaxEntriesInWindow: a.cfg.Analyzer.Cache.RPS * int(a.cfg.Analyzer.RecordTTL.Duration.Seconds()),
		MaxEntrySize:       recSizeTotal,
		HardMaxCacheSize:   int(a.cfg.Analyzer.Cache.MaxSize),
		Verbose:            a.cfg.Log.LogLevel == "debug",
		Logger:             slog.NewLogLogger(a.logger.Handler(), slog.LevelDebug),
		OnRemove:           nil,
		OnRemoveWithReason: nil,
	}
	a.cache, err = bigcache.New(context.Background(), cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return a, nil
}

func (a *Analyzer) ProcessRequest(r dto.Request) {
	a.logger.Debug("processing request", "request", r)

	// Early return for invalid requests
	if r.BodySent <= 0 {
		return
	}

	// Apply filters
	// include first, then exclude
	include := false
	for _, v := range a.cfg.Analyzer.Include {
		if v.Match(r) {
			include = true
			break
		}
	}
	if !include {
		return
	}

	for _, v := range a.cfg.Analyzer.Exclude {
		if v.Match(r) {
			return
		}
	}

	// Update (or create) record
	record, err := a.getRecord(r.Client)
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			record.Addr = r.Client
		} else {
			a.logger.Error("failed to get record", logging.SlogKeyError, err)
			return
		}
	}

	currentTime := time.Now()
	record.Bucket = max(0, record.Bucket-int64(currentTime.Sub(record.LastModified).Seconds())*int64(a.cfg.Analyzer.LeakRate))
	record.Bucket += r.BodySent
	record.LastModified = currentTime
	if record.Bucket > int64(a.cfg.Analyzer.Capacity) {
		prefixLength := 32
		if record.Addr.Is4() {
			prefixLength = a.cfg.Analyzer.DenyPrefixLength.IPv4
		} else if record.Addr.Is6() {
			prefixLength = a.cfg.Analyzer.DenyPrefixLength.IPv6
		}

		err := a.denylist.PutEntry(
			netip.PrefixFrom(record.Addr, prefixLength),
			record.Addr,
		)
		if err != nil {
			a.logger.Error("failed to put denylist entry", logging.SlogKeyError, err)
		}
	}

	err = a.putRecord(record)
	if err != nil {
		a.logger.Error("failed to put record", logging.SlogKeyError, err)
	}
}

func (a *Analyzer) Close() {
	err := a.cache.Close()
	if err != nil {
		a.logger.Error("failed to close cache", logging.SlogKeyError, err)
	}
}
