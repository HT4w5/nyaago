package server

import (
	"os"

	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/aclfmt"
	"github.com/HT4w5/nyaago/pkg/meta"
)

func (s *Server) writeACL() {
	s.logger.Info("writing ACL config")

	// Create formatter
	formatter, err := aclfmt.MakeFormatter(s.cfg.Egress.Format, meta.GetMetadataSingleLine())
	if err != nil {
		s.logger.Error("failed to create formatter", logging.SlogKeyError, err)
	}

	// Open file for write
	f, err := os.OpenFile(s.cfg.Egress.Path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0664)
	if err != nil {
		s.logger.Error("failed to open config file", logging.SlogKeyError, err, "path", s.cfg.Egress.Path)
		return
	}

	rules, err := s.rulelist.ListRules()
	if err != nil {
		s.logger.Error("failed to get rules", logging.SlogKeyError, err)
		return
	}
	err = formatter.Marshal(rules, f)
	if err != nil {
		s.logger.Error("failed to write config", logging.SlogKeyError, err, "path", s.cfg.Egress.Path, "formatter_type", s.cfg.Egress.Format)
		return
	}
}
