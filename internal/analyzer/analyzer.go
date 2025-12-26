package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/iplist"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/allegro/bigcache/v3"
)

const (
	slogModuleName = "analyzer"
	slogGroupName  = "analyzer"
)

const (
	minRateLimit = 100000 // 100kbps
)

type Analyzer struct {
	cfg    *config.Config
	cache  *bigcache.BigCache
	logger *slog.Logger
	iplist *iplist.IPList
}

func MakeAnalyzer(cfg *config.Config, iplist *iplist.IPList) (*Analyzer, error) {
	a := &Analyzer{
		cfg:    cfg,
		iplist: iplist,
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
		MaxEntrySize:       recEncodedSize,
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

	// Drop if too old
	if r.Time.Compare(record.LastModified) <= 0 {
		a.logger.Warn("dropped obsolete request")
		return
	}
	record.Bucket = max(0, record.Bucket-int64(r.Time.Sub(record.LastModified).Seconds())*int64(a.cfg.Analyzer.LeakRate))
	record.Bucket += r.BodySent
	record.LastModified = r.Time
	if record.Bucket > int64(a.cfg.Analyzer.Capacity) {
		prefixLength := 32
		if record.Addr.Is4() {
			prefixLength = a.cfg.Analyzer.LimitPrefixLength.IPv4
		} else if record.Addr.Is6() {
			prefixLength = a.cfg.Analyzer.LimitPrefixLength.IPv6
		}

		// Calculate rate limit
		severity := float64(record.Bucket) / float64(a.cfg.Analyzer.Capacity)
		ratelimit := max(float64(a.cfg.Analyzer.LeakRate)/severity/severity, minRateLimit)

		err := a.iplist.PutRule(dto.Rule{
			Valid:     true,
			Blame:     record.Addr,
			Prefix:    netip.PrefixFrom(record.Addr, prefixLength).Masked(),
			RateLimit: int64(ratelimit),
		})
		if err != nil {
			a.logger.Error("failed to put iplist entry", logging.SlogKeyError, err)
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
