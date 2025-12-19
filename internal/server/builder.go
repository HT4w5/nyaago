package server

import (
	"context"
	"os"

	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/aclfmt"
	"github.com/HT4w5/nyaago/pkg/meta"
)

func (s *Server) runMainCronTask(ctx context.Context) {
	s.runBuildRules(ctx)
	s.runWriteConfig(ctx)
	s.runPostExec(ctx)
}

func (s *Server) runBuildRules(ctx context.Context) {
	// Flush expired
	s.logger.Info("flushing expired pool objects")
	err := s.pool.FlushExpired()
	if err != nil {
		s.logger.Error("failed to flush expired pool objects: %w", err)
	}

	// Build rules
	s.logger.Info("building new rules")
	err = s.pool.BuildRules()
	if err != nil {
		s.logger.Error("failed to build new rules: %w", err)
	}
}

func (s *Server) runWriteConfig(ctx context.Context) {
	// Create formatter
	formatter, err := aclfmt.MakeFormatter(s.cfg.Fmt.Type, meta.GetMetadataSingleLine())
	if err != nil {
		s.logger.Error("failed to create formatter", logging.LoggerKeyError, err)
	}

	// Open file for write
	s.logger.Info("writing ACL config")
	f, err := os.OpenFile(s.cfg.Fmt.Path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0664)
	if err != nil {
		s.logger.Error("failed to open config file", logging.LoggerKeyError, err, "path", s.cfg.Fmt.Path)
		return
	}

	set, err := s.pool.GetRuleSet()
	if err != nil {
		s.logger.Error("failed to get ruleset", logging.LoggerKeyError, err)
		return
	}
	err = formatter.Marshal(set, f)
	if err != nil {
		s.logger.Error("failed to write config", logging.LoggerKeyError, err, "path", s.cfg.Fmt.Path, "formatter_type", s.cfg.Fmt.Type)
		return
	}
}
