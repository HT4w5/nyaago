package analyzer

import (
	"context"
	"log/slog"

	"github.com/HT4w5/nyaago/internal/analyzer/fsr"
	"github.com/HT4w5/nyaago/internal/analyzer/lbucket"
	"github.com/HT4w5/nyaago/internal/analyzer/rfreq"
	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/internal/rulelist"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/dgraph-io/badger/v4"
)

const (
	slogModuleName = "analyzer_manager"
	slogGroupName  = "analyzer_manager"
)

type AnalyzerManager struct {
	cfg       *config.AnaylzerConfig
	db        *badger.DB
	analyzers []Analyzer
	logger    *slog.Logger
}

func MakeAnalyzerManager(cfg *config.AnaylzerConfig, db *badger.DB) *AnalyzerManager {
	am := AnalyzerManager{
		cfg:       cfg,
		db:        db,
		analyzers: make([]Analyzer, 0),
		logger:    logging.GetLogger().With(logging.SlogKeyModule, slogModuleName).WithGroup(slogGroupName),
	}

	// Make analyzers
	if cfg.LeakyBucket.Enabled {
		am.analyzers = append(am.analyzers, lbucket.MakeLeakyBucket(&cfg.LeakyBucket, db))
	}
	if cfg.FileSendRatio.Enabled {
		am.analyzers = append(am.analyzers, fsr.MakeFileSendRatio(&cfg.FileSendRatio, db))
	}
	if cfg.RequestFrequency.Enabled {
		am.analyzers = append(am.analyzers, rfreq.MakeRequestFrequency(&cfg.RequestFrequency, db))
	}

	return &am
}

// Start all enabled analyzers
func (am *AnalyzerManager) Start(ctx context.Context) error {
	enabledAnalyzers := make([]string, 0)
	for _, v := range am.analyzers {
		enabledAnalyzers = append(enabledAnalyzers, v.Name())
	}
	am.logger.Info("starting analyzers", "enabled_analyzers", enabledAnalyzers)

	for _, v := range am.analyzers {
		err := v.Start(ctx)
		if err != nil {
			am.logger.Error("failed to start analyzer", "analyzer", v.Name(), logging.SlogKeyError, err)
			return err
		}
	}
	return nil
}

// Send a request to enabled analyzers
func (am *AnalyzerManager) Process(request dto.Request) {
	am.logger.Debug("processing request", "request", request)
	for _, v := range am.analyzers {
		err := v.Process(request)
		if err != nil {
			am.logger.Error("failed to process request", "analyzer", v.Name(), logging.SlogKeyError, err)
		}
	}
}

// Generate rules from analyzers and save to rulelist
func (am *AnalyzerManager) SaveRules(rl *rulelist.RuleList) {
	tx := rl.BeginTx()
	for _, v := range am.analyzers {
		err := v.Report(tx)
		if err != nil {
			am.logger.Error("analyzer report failed", "analyzer", v.Name(), logging.SlogKeyError, err)
		}
	}
	err := tx.Commit()
	if err != nil {
		am.logger.Error("failed to commit rulelist tx", logging.SlogKeyError, err)
	}
}
