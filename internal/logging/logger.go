package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/gin-gonic/gin"
)

const (
	SlogKeyModule = "module"
	SlogKeyError  = "error"
)

var logger *slog.Logger

func GetLogger(cfg *config.LogConfig) (*slog.Logger, error) {
	if logger != nil {
		return logger, nil
	}

	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	var writer io.Writer
	switch cfg.Access {
	case "none":
		return slog.New(slog.DiscardHandler), nil
	case "":
		writer = os.Stdout
	default:
		f, err := os.OpenFile(cfg.Access, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %s: %w", cfg.Access, err)
		}
		writer = f
	}

	var handler slog.Handler
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	case "none":
		return slog.New(slog.DiscardHandler), nil
	case "":
		level = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid log level: %s", cfg.LogLevel)
	}

	if cfg.Json {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level:     level,
			AddSource: cfg.LogLevel == "debug",
		})
	} else {
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level:     level,
			AddSource: cfg.LogLevel == "debug",
		})
	}

	logger = slog.New(handler)
	return logger, nil
}
