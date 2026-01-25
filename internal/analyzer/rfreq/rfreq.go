package rfreq

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/dbkey"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/internal/rulelist"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/dgraph-io/badger/v4"
)

const (
	analyzerName   = "request_frequency"
	slogModuleName = "rfreq"
	slogGroupName  = "rfreq"
)

type RequestFrequency struct {
	cfg           *config.RequestFrequencyConfig
	db            *badger.DB
	kb            dbkey.KeyBuilder
	logger        *slog.Logger
	reqCountMap   map[netip.Addr]int
	blameTemplate string
}

func MakeRequestFrequency(cfg *config.RequestFrequencyConfig, db *badger.DB) *RequestFrequency {
	kb := dbkey.KeyBuilder{}.WithPrefix(dbkey.RequestFrequency)
	return &RequestFrequency{
		cfg:         cfg,
		db:          db,
		kb:          kb,
		logger:      logging.GetLogger().With(logging.SlogKeyModule, slogModuleName).WithGroup(slogGroupName),
		reqCountMap: make(map[netip.Addr]int),
		blameTemplate: fmt.Sprintf(
			"RPS exceeded %f.",
			cfg.RPSThreshold,
		),
	}
}

func (rf *RequestFrequency) Name() string {
	return analyzerName
}

func (rf *RequestFrequency) Start(ctx context.Context) error {
	// Start timer
	go rf.compileTicker(ctx)
	return nil
}

func (rf *RequestFrequency) Process(request dto.Request) error {
	rf.reqCountMap[request.Client]++
	return nil
}

func (rf *RequestFrequency) Report(tx *rulelist.Tx) error {
	recMap := make(map[netip.Addr]record)
	err := rf.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = rf.kb.Build()
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var rec record
			err := item.Value(func(val []byte) error {
				return rec.Unmarshal(val)
			})
			if err != nil {
				return err
			}
			oldRec, ok := recMap[rec.Addr]
			if !ok {
				if rec.RPS >= rf.cfg.RPSThreshold {
					recMap[rec.Addr] = rec
				}
			} else {
				if rec.RPS > oldRec.RPS {
					recMap[rec.Addr] = rec
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Generate rules
	expTime := time.Now().Add(time.Duration(rf.cfg.Export.TTL))
	for _, v := range recMap {
		// Get prefix
		prefixLength := 32
		if v.Addr.Is4() {
			prefixLength = rf.cfg.Export.PrefixLength.IPv4
		} else if v.Addr.Is6() {
			prefixLength = rf.cfg.Export.PrefixLength.IPv6
		}
		prefix := netip.PrefixFrom(v.Addr, prefixLength).Masked()

		tx.PutRule(dto.Rule{
			Prefix: prefix,
			Banned: true,
			Blame: fmt.Sprintf(
				"%s Actual RPS %.2f.",
				rf.blameTemplate,
				v.RPS,
			),
			ExpiresAt: expTime,
		})
	}

	return nil
}

func (rf *RequestFrequency) compileRecords() {
	recs := make([]record, 0, len(rf.reqCountMap))
	for k, v := range rf.reqCountMap {
		rec := record{
			Addr:     k,
			RPS:      float64(v) / float64(rf.cfg.UnitTime),
			Duration: time.Duration(rf.cfg.UnitTime),
		}
		recs = append(recs, rec)
	}

	err := rf.putRecords(recs)
	if err != nil {
		rf.logger.Error("failed to put records", logging.SlogKeyError, err)
	}

	// Clean-up
	clear(rf.reqCountMap)
}

func (rf *RequestFrequency) compileTicker(ctx context.Context) {
	rf.logger.Info("starting compile ticker")
	ticker := time.NewTicker(time.Duration(rf.cfg.UnitTime))
Loop:
	for {
		select {
		case <-ticker.C:
			rf.compileRecords()
		case <-ctx.Done():
			break Loop
		}
	}
	rf.logger.Info("stopping compile ticker")
}
