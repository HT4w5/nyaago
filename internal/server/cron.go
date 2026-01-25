package server

import (
	"context"
	"time"

	"github.com/go-co-op/gocron/v2"
)

func (s *Server) setupCronJobs() {
	s.cron.NewJob(
		gocron.DurationJob(time.Duration(s.cfg.Egress.Interval)),
		gocron.NewTask(s.runEgressTask),
	)
}

func (s *Server) runEgressTask(ctx context.Context) {
	s.writeACL()
	s.postExec(ctx)
}
