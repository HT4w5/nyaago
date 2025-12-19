package tail

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/HT4w5/nyaago/pkg/parser"
	"github.com/nxadm/tail"
	tailutil "github.com/nxadm/tail"
)

const (
	loggerModuleName = "tail"
	loggerGroupName  = "tail"
)

type Tail struct {
	cfg    *config.TailConfig
	parser parser.Parser
	tail   *tailutil.Tail
	logger *slog.Logger
}

func MakeTail(cfg *config.TailConfig, logger *slog.Logger) (*Tail, error) {
	t := &Tail{}

	// Setup parser
	switch cfg.Type {
	case "nginxjson":
		var err error
		t.parser, err = parser.MakeParser(cfg.Type)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported log type: %s", cfg.Type)
	}

	// Setup logger
	t.logger = logger.With(logging.LoggerKeyModule, loggerModuleName).WithGroup(loggerGroupName)

	// Setup tail
	var err error
	t.tail, err = tail.TailFile(cfg.Path, tailutil.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Poll:      cfg.Poll,
		Logger:    slog.NewLogLogger(t.logger.Handler(), slog.LevelInfo),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tail: %w", err)
	}

	return t, nil
}

func (t *Tail) Start(ctx context.Context, out chan<- dto.Request, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			t.tail.Stop()
			return
		case line := <-t.tail.Lines:
			if line == nil {
				// Tail failed, cancel global context
				err := t.tail.Wait()
				t.logger.Error("tail failed", logging.LoggerKeyError, err)
				cancel()
				return
			}
			t.logger.Debug("line received", "content", line.Text)
			req, err := t.parser.Parse([]byte(line.Text))
			if err != nil {
				t.logger.Error("failed to parse line", "log_type", t.cfg.Type, "log_path", t.cfg.Path, "log_line", line)
				continue
			}
			out <- req
		}
	}
}

func (t *Tail) Close() {
	t.logger.Info("shutting down tail")
	t.tail.Cleanup()
}
