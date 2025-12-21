package ingress

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/HT4w5/nyaago/pkg/parser"
)

const (
	slogModuleName = "ingress"
	slogGroupName  = "ingress"

	slogKeyLogFormat = "log_format"
	slogKeySource    = "source"
	slogKeyMethod    = "method"
	slogKeyLine      = "line"
)

type IngressAdapter interface {
	Start(ctx context.Context, out chan<- dto.Request, cancel context.CancelFunc)
	Close()
}

func MakeIngressAdapter(cfg *config.IngressConfig, logger *slog.Logger) (IngressAdapter, error) {
	// Setup parser
	var p parser.Parser
	switch cfg.Format {
	case "nginxjson":
		var err error
		p, err = parser.MakeParser(cfg.Format)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported ingress log format: %s", cfg.Format)
	}

	// Setup logger
	logger = logger.With(logging.SlogKeyModule, slogModuleName).WithGroup(slogModuleName)

	switch cfg.Method {
	case "tail":
		ti, err := makeTailIngress(cfg, p, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create tail ingress: %w", err)
		}
		return ti, nil
	default:
		return nil, fmt.Errorf("unsupported ingress method: %s", cfg.Method)
	}
}
