package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/ingress"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/db"
	"github.com/go-co-op/gocron/v2"
)

const (
	slogModuleNameServer = "server"
	slogModuleNameCron   = "cron"

	slogGroupNameServer = "server"
	slogGroupNameCron   = "cron"
)

type Server struct {
	cfg    *config.Config
	db     db.DBAdapter
	ia     ingress.IngressAdapter
	cron   gocron.Scheduler
	logger *slog.Logger
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
	logger, err := logging.GetLogger(&cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	// Make DBAdapter
	s.db, err = db.MakeDBAdapter(cfg.DB.Type, cfg.DB.Access)
	if err != nil {
		return nil, fmt.Errorf("failed to get DBAdapter: %w", err)
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
	s.ia, err = ingress.MakeIngressAdapter(&cfg.Ingress, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create ingress adapter: %w", err)
	}

	s.logger = logger.With(logging.SlogKeyModule, slogModuleNameServer).WithGroup(slogGroupNameServer)

	server = s
	return s, nil
}

func (s *Server) Start(ctx context.Context, cancel context.CancelFunc) {
	s.logger.Info("starting")
	s.logger.Info("db driver info", "db_info", s.db.Info())
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
	s.logger.Info("exiting")
}
