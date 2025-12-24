package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/internal/server"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
)

type API struct {
	engine *gin.Engine
	http   *http.Server
	srv    *server.Server
	logger *slog.Logger
}

const (
	slogModuleName = "api"
	slogGroupName  = "api"

	slogModuleNameGin = "gin"
	slogGroupNameGin  = "gin"
)

func MakeAPI(cfg *config.Config, s *server.Server) (*API, error) {
	logger, err := logging.GetLogger(&cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}
	api := &API{
		engine: gin.New(),
		logger: logger.With(logging.SlogKeyModule, slogModuleName).WithGroup(slogGroupName),
		srv:    s,
	}

	// Setup gin router
	api.engine.Use(sloggin.NewWithConfig(logger.With(logging.SlogKeyModule, slogModuleNameGin).WithGroup(slogGroupNameGin), sloggin.Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,
		HandleGinDebug:   true,
	}))
	api.engine.Use(gin.Recovery())
	api.setupRoutesV1()

	// Setup HTTP server
	api.http = &http.Server{
		Addr:    cfg.API.ListenAddr,
		Handler: api.engine,
	}

	return api, nil
}

func (api *API) Start() {
	api.logger.Info("starting")
	// HTTP server
	go func() {
		if err := api.http.ListenAndServe(); err != http.ErrServerClosed && err != nil {
			api.logger.Error("listen failed", logging.SlogKeyError, err)
		}
	}()
}

func (api *API) Shutdown(ctx context.Context) {
	api.logger.Info("shutting down")
	err := api.http.Shutdown(ctx)
	if err != nil {
		api.logger.Error("failed to shutdown HTTP server", logging.SlogKeyError, err)
	}
}
