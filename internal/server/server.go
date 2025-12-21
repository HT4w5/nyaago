package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/ingress"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/internal/pool"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	sloggin "github.com/samber/slog-gin"
)

const (
	slogModuleNameServer = "server"
	slogModuleNameCron   = "cron"

	slogGroupNameServer = "server"
	slogGroupNameCron   = "cron"
)

type Server struct {
	cfg    *config.Config
	router *gin.Engine
	srv    *http.Server
	pool   *pool.Pool
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
		cfg:    cfg,
		router: gin.New(),
	}

	var err error
	// Create logger
	logger, err := logging.SetupLogger(&cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Create cron scheduler
	s.cron, err = gocron.NewScheduler(
		gocron.WithLogger(logger.With(logging.SlogKeyModule, slogModuleNameCron).WithGroup(slogGroupNameCron)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron scheduler: %w", err)
	}
	s.setupCronJobs()

	// Setup gin router
	s.router.Use(sloggin.NewWithConfig(logger, sloggin.Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,
	}))
	s.router.Use(gin.Recovery())
	s.setupRoutes()

	// Setup HTTP server
	s.srv = &http.Server{
		Addr:    cfg.API.ListenAddr,
		Handler: s.router,
	}

	// Create pool
	s.pool, err = pool.GetPool(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

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
	s.logger.Info("starting server")
	// Cron
	s.cron.Start()

	// Ingress worker
	go s.runIngressWorker(ctx, cancel)

	// HTTP server
	go func() {
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed && err != nil {
			s.logger.Error("listen failed", logging.SlogKeyError, err)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) {
	s.logger.Info("shutting down")
	err := s.srv.Shutdown(ctx)
	if err != nil {
		s.logger.Error("failed to shutdown HTTP server", logging.SlogKeyError, err)
	}
	err = s.cron.Shutdown()
	if err != nil {
		s.logger.Error("failed to shutdown gocron scheduler", logging.SlogKeyError, err)
	}
	s.logger.Info("exiting")
}
