package fsr

import (
	"context"
	"log/slog"

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
		crKb:   kb.WithPrefix(currentRecords),
		hrKb:   kb.WithPrefix(historicRecords),
		fsrKb:  kb.WithPrefix(fileSizeRecords),
		logger: logging.GetLogger().With(logging.SlogKeyModule, slogModuleName).WithGroup(slogGroupName),
	}
}

func (fsr *FileSendRatio) Start(ctx context.Context) error {

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
	return fsr.putCurrentRecord(rec)
}

func (fsr *FileSendRatio) Report(tx *rulelist.Tx) error {

}
