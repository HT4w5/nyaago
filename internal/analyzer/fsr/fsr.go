package fsr

import (
	"context"
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
	slogModuleName = "fsr"
	slogGroupName  = "fsr"
)

type FileSendRatio struct {
	cfg    *config.FileSendRatioConfig
	db     *badger.DB
	kb     dbkey.KeyBuilder // Main key builder
	ipKb   dbkey.KeyBuilder // for ip records
	crKb   dbkey.KeyBuilder // for current records
	hrKb   dbkey.KeyBuilder // for historic records
	fsrKb  dbkey.KeyBuilder // for file size records
	logger *slog.Logger
}

func MakeFileSendRatio(cfg *config.FileSendRatioConfig, db *badger.DB) *FileSendRatio {
	kb := dbkey.KeyBuilder{}.WithPrefix(dbkey.FileSendRatio)
	return &FileSendRatio{
		cfg:    cfg,
		db:     db,
		kb:     kb,
		ipKb:   kb.WithPrefix(ipRecords),
		crKb:   kb.WithPrefix(currentRecords),
		hrKb:   kb.WithPrefix(historicRecords),
		fsrKb:  kb.WithPrefix(fileSizeRecords),
		logger: logging.GetLogger().With(logging.SlogKeyModule, slogModuleName).WithGroup(slogGroupName),
	}
}

func (fsr *FileSendRatio) Start(ctx context.Context) error {
	// Clean-up previous
	err := fsr.clearCurrentRecords()
	if err != nil {
		return err
	}
	err = fsr.clearIPRecords()
	if err != nil {
		return err
	}

	// Start timer
	go fsr.compileTicker(ctx)
	return nil
}

func (fsr *FileSendRatio) Process(request dto.Request) error {
	rec, err := fsr.getCurrentRecord(request.Client, request.URL)
	if err != nil {
		if err == ErrRecordNotFound {
			rec = currentRecord{
				Addr: request.Client,
				Path: request.URL,
				Sent: 0,
			}
		} else {
			return err
		}
	}

	rec.Sent += request.Sent
	err = fsr.putIPRecord(ipRecord{
		Addr: request.Client,
	})
	if err != nil {
		return err
	}
	return fsr.putCurrentRecord(rec)
}

// Lookup all historic records, find greatest ratio for each IP,
// and report those over the threshold
func (fsr *FileSendRatio) Report(tx *rulelist.Tx) error {
	recMap := make(map[netip.Addr]historicRecord)
	err := fsr.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = fsr.hrKb.Build()
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var rec historicRecord
			err := item.Value(func(val []byte) error {
				return rec.Unmarshal(val)
			})
			if err != nil {
				return err
			}
			oldRec, ok := recMap[rec.Addr]
			if !ok {
				if rec.Ratio >= fsr.cfg.Export.RatioThreshold {
					recMap[rec.Addr] = rec
				}
			} else {
				if rec.Ratio > oldRec.Ratio {
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
	expTime := time.Now().Add(time.Duration(fsr.cfg.Export.TTL))
	for _, v := range recMap {
		// Get prefix
		prefixLength := 32
		if v.Addr.Is4() {
			prefixLength = fsr.cfg.Export.PrefixLength.IPv4
		} else if v.Addr.Is6() {
			prefixLength = fsr.cfg.Export.PrefixLength.IPv6
		}
		prefix := netip.PrefixFrom(v.Addr, prefixLength).Masked()

		tx.PutRule(dto.Rule{
			Prefix:    prefix,
			Banned:    true,
			Blame:     v.blame(),
			ExpiresAt: expTime,
		})
	}

	return nil
}

func (fsr *FileSendRatio) compileTicker(ctx context.Context) {
	fsr.logger.Info("starting compile ticker")
	ticker := time.NewTicker(time.Duration(fsr.cfg.UnitTime))
Loop:
	for {
		select {
		case <-ticker.C:
			fsr.compileHistoricRecords()
		case <-ctx.Done():
			break Loop
		}
	}
	fsr.logger.Info("stopping compile ticker")
}

// Summarize current records and create new historic records
func (fsr *FileSendRatio) compileHistoricRecords() {
	ipRecs, err := fsr.getAllIPRecords()
	hisRecs := make([]historicRecord, 0, len(ipRecs))
	if err != nil {
		fsr.logger.Error("failed to get ip records", logging.SlogKeyError, err)
	}

	for _, ip := range ipRecs {
		maxRatio := -1.
		var maxRec currentRecord
		err := fsr.db.View(func(txn *badger.Txn) error {
			opt := badger.DefaultIteratorOptions
			opt.Prefix = fsr.hrKb.WithObject(ip).Build()
			it := txn.NewIterator(opt)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				var rec currentRecord
				err := item.Value(func(val []byte) error {
					return rec.Unmarshal(val)
				})
				if err != nil {
					return err
				}
				// Calculate ratio
				size, ok := fsr.getPathSize(rec.Path)
				if !ok {
					continue
				}
				ratio := float64(rec.Sent) / float64(size)
				if ratio > maxRatio {
					maxRatio = ratio
					maxRec = rec
				}
			}
			return nil
		})
		if err != nil {
			fsr.logger.Error("failed to create historic record", logging.SlogKeyError, err, "addr", ip.Addr)
			continue
		}

		if maxRatio < 0. {
			fsr.logger.Warn("no valid current record, skipping address", "addr", ip.Addr)
			continue
		}

		hisRec := historicRecord{
			Addr:     ip.Addr,
			Path:     maxRec.Path,
			Ratio:    maxRatio,
			Time:     time.Now(),
			Duration: time.Duration(fsr.cfg.UnitTime),
		}

		hisRecs = append(hisRecs, hisRec)
	}

	err = fsr.putHistoricRecords(hisRecs)
	if err != nil {
		fsr.logger.Error("failed to put historic records", logging.SlogKeyError, err)
	}

	// Clean up current records
	err = fsr.clearCurrentRecords()
	if err != nil {
		fsr.logger.Error("failed to clear current records", logging.SlogKeyError, err)
	}
}
