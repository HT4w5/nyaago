package server

import (
	"context"

	"github.com/go-co-op/gocron/v2"
)

func (s *Server) setupCronJobs() {
	s.cron.NewJob(
		gocron.DurationJob(s.cfg.Egress.Interval.Duration),
		gocron.NewTask(s.runEgressTask),
	)
}

func (s *Server) runEgressTask(ctx context.Context) {
	s.writeACL()
	s.postExec(ctx)
}
