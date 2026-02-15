package server

import (
	"context"
)

func (s *Server) runTicker(ctx context.Context) {
	s.writeACL()
	s.postExec(ctx)
}
