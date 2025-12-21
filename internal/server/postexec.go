package server

import (
	"context"
	"os/exec"

	"github.com/HT4w5/nyaago/internal/logging"
)

func (s *Server) runPostExec(ctx context.Context) {
	total := len(s.cfg.Egress.PostExec)
	for i, v := range s.cfg.Egress.PostExec {
		s.logger.Info("running postexec", "total", total, "current", i, "tag", v.Tag)

		cmd := exec.CommandContext(ctx, v.Cmd, v.Args...)
		if v.Cwd != "" {
			cmd.Dir = v.Cwd
		}

		err := cmd.Run()
		if err != nil {
			s.logger.Error("failed to run postexec", "tag", v.Tag, logging.SlogKeyError, err)
		}
	}
}
