package ingress

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/HT4w5/nyaago/pkg/parser"
	"github.com/nxadm/tail"
	tailutil "github.com/nxadm/tail"
)

type TailIngress struct {
	cfg    *config.IngressConfig
	parser parser.Parser
	tail   *tailutil.Tail
	logger *slog.Logger
}

func makeTailIngress(cfg *config.IngressConfig, p parser.Parser, logger *slog.Logger) (*TailIngress, error) {
	i := &TailIngress{
		parser: p,
		logger: logger,
	}

	// Setup tail
	var err error
	i.tail, err = tail.TailFile(cfg.Tail.Path, tailutil.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Poll:      cfg.Tail.Poll,
		Logger:    slog.NewLogLogger(i.logger.Handler(), slog.LevelInfo),
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tail: %w", err)
	}

	return i, nil
}

func (i *TailIngress) Start(ctx context.Context, out chan<- dto.Request, cancel context.CancelFunc) {
	i.logger.Info("starting tail")
	for {
		select {
		case <-ctx.Done():
			i.logger.Info("shutting down tail")
			i.tail.Stop()
			i.tail.Cleanup()
			return
		case line := <-i.tail.Lines:
			if line == nil {
				// Tail failed, cancel global context
				err := i.tail.Wait()
				i.logger.Error("tail failed", logging.SlogKeyError, err)
				cancel()
				return
			}
			i.logger.Debug("line received", "content", line.Text)
			req, err := i.parser.Parse([]byte(line.Text))
			if err != nil {
				i.logger.Error("failed to parse line", slogKeyMethod, i.cfg.Method, slogKeyLogFormat, i.cfg.Format, slogKeySource, i.cfg.Tail.Path, slogKeyLine, line.Text)
				continue
			}
			out <- req
		}
	}
}
