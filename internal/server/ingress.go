package server

import (
	"context"

	"github.com/HT4w5/nyaago/pkg/dto"
)

func (s *Server) runIngressWorker(ctx context.Context, cancel context.CancelFunc) {
	s.logger.Info("starting ingress worker")
	requestChan := make(chan dto.Request)

	// Start ingress adapter
	go s.ia.Start(ctx, requestChan, cancel)

	// Worker loop
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-requestChan:
			s.router.ProcessRequest(req)
		}
	}
}
