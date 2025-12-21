package server

import (
	"context"
	"os/exec"

	"github.com/HT4w5/nyaago/internal/logging"
)

func (s *Server) runPostExec(ctx context.Context) {
	cmd := exec.CommandContext(ctx, s.cfg.Egress.PostExec.Cmd, s.cfg.Egress.PostExec.Args...)
	if s.cfg.Egress.PostExec.Cwd != "" {
		cmd.Dir = s.cfg.Egress.PostExec.Cwd
	}

	err := cmd.Run()
	if err != nil {
		s.logger.Error("failed to run postexec", logging.LoggerKeyError, err)
	}
}
