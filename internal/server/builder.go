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
		s.logger.Error("failed to flush expired pool objects", logging.SlogKeyError, err)
	}

	// Build rules
	s.logger.Info("building new rules")
	err = s.pool.BuildRules()
	if err != nil {
		s.logger.Error("failed to build new rules", logging.SlogKeyError, err)
	}
}

func (s *Server) runWriteConfig(ctx context.Context) {
	// Create formatter
	formatter, err := aclfmt.MakeFormatter(s.cfg.Egress.Format, meta.GetMetadataSingleLine())
	if err != nil {
		s.logger.Error("failed to create formatter", logging.SlogKeyError, err)
	}

	// Open file for write
	s.logger.Info("writing ACL config")
	f, err := os.OpenFile(s.cfg.Egress.Path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0664)
	if err != nil {
		s.logger.Error("failed to open config file", logging.SlogKeyError, err, "path", s.cfg.Egress.Path)
		return
	}

	set, err := s.pool.GetRuleSet()
	if err != nil {
		s.logger.Error("failed to get ruleset", logging.SlogKeyError, err)
		return
	}
	err = formatter.Marshal(set, f)
	if err != nil {
		s.logger.Error("failed to write config", logging.SlogKeyError, err, "path", s.cfg.Egress.Path, "formatter_type", s.cfg.Egress.Format)
		return
	}
}
