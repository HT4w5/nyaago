package server

import "github.com/go-co-op/gocron/v2"

func (s *Server) setupCronJobs() {
	s.cron.NewJob(
		gocron.DurationJob(s.cfg.Cron.Interval.Duration),
		gocron.NewTask(s.runMainCronTask),
	)
}
