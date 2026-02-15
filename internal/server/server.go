package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/HT4w5/nyaago/internal/analyzer"
	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/ingress"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/internal/rulelist"
	"github.com/dgraph-io/badger/v4"
)

const (
	slogModuleNameServer = "server"
	slogModuleNameCron   = "cron"

	slogGroupNameServer = "server"
	slogGroupNameCron   = "cron"
)

type Server struct {
	cfg      *config.Config
	db       *badger.DB
	rulelist *rulelist.RuleList
	ia       ingress.IngressAdapter
	am       *analyzer.AnalyzerManager
	logger   *slog.Logger
}

var server *Server

func GetServer(cfg *config.Config) (*Server, error) {
	if server != nil {
		return server, nil
	}

	s := &Server{
		cfg: cfg,
	}

	var err error
	// Create logger
	logger := logging.GetLogger()

	// Open DB
	s.db, err = badger.Open(badger.DefaultOptions(s.cfg.DB.Dir))
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// Create AnalyzerManager
	s.am = analyzer.MakeAnalyzerManager(&cfg.Anaylzer, s.db)

	// Create RuleList
	s.rulelist, err = rulelist.MakeRuleList(cfg, s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create rulelist: %w", err)
	}

	// Create ingress adapter
	s.ia, err = ingress.MakeIngressAdapter(&cfg.Ingress)
	if err != nil {
		return nil, fmt.Errorf("failed to create ingress adapter: %w", err)
	}

	s.logger = logger.With(logging.SlogKeyModule, slogModuleNameServer).WithGroup(slogGroupNameServer)

	server = s
	return s, nil
}

func (s *Server) Start(ctx context.Context, cancel context.CancelFunc) {
	s.logger.Info("starting")

	// Create egress file
	s.writeACL()

	// Ingress worker
	go s.runIngressWorker(ctx, cancel)
}

func (s *Server) Shutdown(ctx context.Context) {
	s.logger.Info("shutting down")

	s.db.Close()

	s.logger.Info("exiting")
}
