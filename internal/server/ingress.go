package server

import (
	"context"

	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/dto"
)

func (s *Server) runIngressWorker(ctx context.Context, cancel context.CancelFunc) {
	requestChan := make(chan dto.Request)

	// Start tail worker
	go s.tail.Start(ctx, requestChan, cancel)

	// Worker loop
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-requestChan:
			err := s.pool.ProcessRequest(req)
			if err != nil {
				s.logger.Error("failed to process request", logging.LoggerKeyError, err)
			}
		}
	}
}
