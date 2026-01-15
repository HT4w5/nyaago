package main

import (
	"context"
	"flag" // Added
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HT4w5/nyaago/internal/api"
	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/internal/server"
	"github.com/HT4w5/nyaago/pkg/meta"
)

const (
	exitSuccess = iota
	exitConfigError
	exitLoggerError
	exitServerError
	exitAPIError
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "config.json", "path to the configuration file")
	flag.StringVar(&cfgPath, "c", "config.json", "path to the configuration file (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s", meta.GetMetadataMultiline())
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	fmt.Print(meta.GetMetadataMultiline())

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Printf("Failed to load config %s: %v\n", cfgPath, err)
		os.Exit(exitConfigError)
	}

	err = logging.Init(&cfg.Log)
	if err != nil {
		fmt.Printf("Failed to setup logger: %v\n", err)
		os.Exit(exitLoggerError)
	}

	srv, err := server.GetServer(cfg)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		os.Exit(exitServerError)
	}

	api, err := api.MakeAPI(cfg, srv)
	if err != nil {
		fmt.Printf("Failed to create api: %v\n", err)
		os.Exit(exitAPIError)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	srv.Start(ctx, stop)
	api.Start()

	<-ctx.Done()
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	api.Shutdown(ctx)
	srv.Shutdown(ctx)

	os.Exit(exitSuccess)
}
