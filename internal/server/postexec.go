package server

import (
	"context"
	"os/exec"

	"github.com/HT4w5/nyaago/internal/logging"
)

func (s *Server) runPostExec(ctx context.Context) {
	cmd := exec.CommandContext(ctx, s.cfg.PostExec.Cmd, s.cfg.PostExec.Args...)
	if s.cfg.PostExec.Cwd != "" {
		cmd.Dir = s.cfg.PostExec.Cwd
	}

	err := cmd.Run()
	if err != nil {
		s.logger.Error("failed to run postexec", logging.LoggerKeyError, err)
	}
}
