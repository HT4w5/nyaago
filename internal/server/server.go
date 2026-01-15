package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/ingress"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/internal/router"
	"github.com/HT4w5/nyaago/internal/rulelist"
	"github.com/dgraph-io/badger/v4"
	"github.com/go-co-op/gocron/v2"
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
	router   *router.Router
	rulelist *rulelist.RuleList
	ia       ingress.IngressAdapter
	cron     gocron.Scheduler
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

	// Create RuleList
	s.rulelist, err = rulelist.MakeRuleList(cfg, s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create rulelist: %w", err)
	}

	// Create Router
	s.router, err = router.MakeRouter(&cfg.Router)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	// Create cron scheduler
	s.cron, err = gocron.NewScheduler(
		gocron.WithLogger(logger.With(logging.SlogKeyModule, slogModuleNameCron).WithGroup(slogGroupNameCron)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron scheduler: %w", err)
	}
	s.setupCronJobs()

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
	// Cron
	s.cron.Start()

	// Ingress worker
	go s.runIngressWorker(ctx, cancel)
}

func (s *Server) Shutdown(ctx context.Context) {
	s.logger.Info("shutting down")

	err := s.cron.Shutdown()
	if err != nil {
		s.logger.Error("failed to shutdown gocron scheduler", logging.SlogKeyError, err)
	}

	s.router.Close()
	s.db.Close()

	s.logger.Info("exiting")
}
