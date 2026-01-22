package lbucket

import (
	"fmt"
	"net/netip"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/dbkey"
	"github.com/HT4w5/nyaago/internal/rulelist"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/dgraph-io/badger/v4"
)

type LeakyBucket struct {
	cfg         *config.LeakyBucketConfig
	db          *badger.DB
	kb          *dbkey.KeyBuilder
	cachedRules []dto.Rule
}

func MakeLeakyBucket(cfg *config.LeakyBucketConfig, db *badger.DB) *LeakyBucket {
	kb := (&dbkey.KeyBuilder{}).WithTag(dbkey.LeakyBucketTag)
	return &LeakyBucket{
		cfg:         cfg,
		db:          db,
		kb:          kb,
		cachedRules: make([]dto.Rule, 0),
	}
}

func (lb *LeakyBucket) Process(request dto.Request) error {
	// Early return for invalid requests
	if request.Sent <= 0 {
		return nil
	}

	// Update (or create) record
	rec, err := lb.getRecord(request.Client)
	if err != nil {
		if err == ErrRecordNotFound {
			rec.Addr = request.Client
		} else {
			return fmt.Errorf("failed to load record from database %w", err)
		}
	}

	// Skip bucket and time update if older than last processed request
	if request.Time.Compare(rec.LastModified) > 0 {
		rec.LastModified = request.Time
		rec.Bucket = max(0, rec.Bucket-int64(request.Time.Sub(rec.LastModified).Seconds())*int64(lb.cfg.LeakRate))
	}

	// Add record to cache if condition satisfies
	if rec.Bucket > int64(lb.cfg.Capacity) {
		// Calculate rate limit
		severity := float64(rec.Bucket) / float64(lb.cfg.Capacity)
		ratelimit := max(float64(lb.cfg.LeakRate)/severity/severity, float64(lb.cfg.MinRate))
		// Get prefix
		prefixLength := 32
		if rec.Addr.Is4() {
			prefixLength = lb.cfg.Export.PrefixLength.IPv4
		} else if rec.Addr.Is6() {
			prefixLength = lb.cfg.Export.PrefixLength.IPv6
		}
		prefix := netip.PrefixFrom(rec.Addr, prefixLength).Masked()

		lb.cachedRules = append(lb.cachedRules, dto.Rule{
			Prefix:    prefix,
			Banned:    false,
			RateLimit: int64(ratelimit),
		})
	}

	err = lb.putRecord(rec)
	if err != nil {
		return fmt.Errorf("failed to put record %w", err)
	}

	return nil
}

func (lb *LeakyBucket) Report(tx *rulelist.Tx) error {
	for _, v := range lb.cachedRules {
		err := tx.PutRule(v)
		if err != nil {
			return err
		}
	}
	lb.cachedRules = make([]dto.Rule, 0)
	return nil
}
