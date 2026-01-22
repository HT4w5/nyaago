package router

import (
	"log/slog"

	"github.com/HT4w5/nyaago/internal/analyzer"
	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/logging"
)

const (
	slogModuleName = "router"
	slogGroupName  = "router"
)

const (
	minRateLimit = 100000 // 100kbps
)

type Router struct {
	cfg       *config.RouterConfig
	flow      flow
	analyzers map[string]analyzer.Analyzer
	logger    *slog.Logger
}

func MakeRouter(cfg *config.Config) (*Router, error) {
	r := Router{}

	r.logger = logging.GetLogger().With(logging.SlogKeyModule, slogModuleName).WithGroup(slogGroupName)
}
