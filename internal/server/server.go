package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/internal/pool"
	"github.com/HT4w5/nyaago/internal/tail"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	sloggin "github.com/samber/slog-gin"
)

const (
	loggerModuleNameServer = "server"
	loggerModuleNameCron   = "cron"

	loggerGroupNameServer = "server"
	loggerGroupNameCron   = "cron"
)

type Server struct {
	cfg    *config.Config
	router *gin.Engine
	srv    *http.Server
	pool   *pool.Pool
	tail   *tail.Tail
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
		gocron.WithLogger(logger.With(logging.LoggerKeyModule, loggerModuleNameCron).WithGroup(loggerGroupNameCron)),
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
		Addr: fmt.Sprintf(
			"%s:%d",
			s.cfg.API.Addr,
			s.cfg.API.Port,
		),
		Handler: s.router,
	}

	// Create pool
	s.pool, err = pool.GetPool(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Create tail
	s.tail, err = tail.MakeTail(&cfg.Tail, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create tail: %w", err)
	}

	s.logger = logger.With(logging.LoggerKeyModule, loggerModuleNameServer).WithGroup(loggerGroupNameServer)

	server = s
	return s, nil
}

func (s *Server) Start(ctx context.Context) {
	// Cron
	s.cron.Start()

	// Ingress worker
	go s.runIngressWorker(ctx)

	// HTTP server
	go func() {
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed && err != nil {
			s.logger.Error("listen failed", logging.LoggerKeyError, err)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) {
	s.logger.Info("shutting down")
	err := s.srv.Shutdown(ctx)
	if err != nil {
		s.logger.Error("failed to shutdown HTTP server", logging.LoggerKeyError, err)
	}
	err = s.cron.Shutdown()
	if err != nil {
		s.logger.Error("failed to shutdown gocron scheduler", logging.LoggerKeyError, err)
	}
	s.logger.Info("exiting")
}
