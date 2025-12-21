package ingress

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/HT4w5/nyaago/pkg/parser"
	"gopkg.in/mcuadros/go-syslog.v2"
)

const (
	logPartKeyHostname = "hostname"
	logPartKeyMessage  = "message"
	logPartKeyContent  = "content"
)

type SyslogIngress struct {
	cfg    *config.IngressConfig
	parser parser.Parser
	srv    *syslog.Server
	out    syslog.LogPartsChannel
	logger *slog.Logger
}

func makeSyslogIngress(cfg *config.IngressConfig, p parser.Parser, logger *slog.Logger) (*SyslogIngress, error) {
	i := &SyslogIngress{
		cfg:    cfg,
		parser: p,
		logger: logger,
		out:    make(syslog.LogPartsChannel),
	}

	// Setup syslog server
	handler := syslog.NewChannelHandler(i.out)
	i.srv = syslog.NewServer()
	i.srv.SetFormat(syslog.Automatic)
	i.srv.SetHandler(handler)
	switch cfg.Syslog.Transport {
	case "tcp":
		i.srv.ListenTCP(cfg.Syslog.ListenAddr)
	case "udp":
		i.srv.ListenUDP(cfg.Syslog.ListenAddr)
	case "unixgram":
		i.srv.ListenUnixgram(cfg.Syslog.ListenAddr)
	default:
		return nil, fmt.Errorf("unsupported syslog transport: %s", cfg.Syslog.Transport)
	}

	return i, nil
}

func (i *SyslogIngress) Start(ctx context.Context, out chan<- dto.Request, cancel context.CancelFunc) {
	// Start syslog server
	i.logger.Info("starting syslog server")
	err := i.srv.Boot()
	if err != nil {
		i.logger.Error("failed to start syslog server", logging.SlogKeyError, err)
		cancel()
		return
	}
	i.logger.Info(fmt.Sprintf("syslog server listening at %s://%s", i.cfg.Syslog.Transport, i.cfg.Syslog.ListenAddr))

	for {
		select {
		case <-ctx.Done():
			i.logger.Info("shutting down syslog server")
			if err := i.srv.Kill(); err != nil {
				i.logger.Error("failed to kill syslog server")
			}
			return
		case logPart := <-i.out:
			if logPart == nil {
				err := i.srv.GetLastError()
				i.logger.Error("syslog server died", logging.SlogKeyError, err)
				cancel()
				return
			}
			hostname, ok := logPart[logPartKeyHostname].(string)
			if !ok {
				i.logger.Warn("hostname field missing")
				hostname = ""
			}
			line, ok := logPart[logPartKeyContent].(string)
			if !ok {
				line, ok = logPart[logPartKeyMessage].(string)
				if !ok {
					i.logger.Error("failed to unpack line, couldn't match field")
					continue
				}
			}
			i.logger.Debug("line received", "content", line)
			req, err := i.parser.Parse([]byte(line))
			if err != nil {
				i.logger.Error("failed to parse line", slogKeyMethod, i.cfg.Method, slogKeyLogFormat, i.cfg.Format, slogKeySource, hostname, slogKeyLine, line)
				continue
			}
			out <- req
		}
	}
}
