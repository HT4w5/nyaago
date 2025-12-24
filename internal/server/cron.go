package server

import (
	"context"

	"github.com/go-co-op/gocron/v2"
)

func (s *Server) setupCronJobs() {
	s.cron.NewJob(
		gocron.DurationJob(s.cfg.Analyzer.UpdateInterval.Duration),
		gocron.NewTask(s.runMainCronTask),
	)
}

func (s *Server) runMainCronTask(ctx context.Context) {
	s.flushExpired()
	s.computeRules()
	s.writeACL()
	s.postExec(ctx)
}
