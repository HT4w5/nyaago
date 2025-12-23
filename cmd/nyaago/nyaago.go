package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HT4w5/nyaago/internal/api"
	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/server"
	"github.com/HT4w5/nyaago/pkg/meta"
)

const (
	exitSuccess = iota
	exitConfigError
	exitServerError
	exitAPIError
)

func main() {
	// Print Metadata
	fmt.Println(meta.GetMetadataMultiline())

	// Load config
	cfgPath := "config.json"
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Printf("Failed to load config %s: %v", cfgPath, err)
		os.Exit(exitConfigError)
	}

	// Create server
	srv, err := server.GetServer(cfg)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		os.Exit(exitServerError)
	}

	// Create API
	api, err := api.MakeAPI(cfg, srv)
	if err != nil {
		fmt.Printf("Failed to create api: %v", err)
		os.Exit(exitAPIError)
	}

	// Listen for interrupt
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// Start server
	srv.Start(ctx, stop)

	// Start api
	api.Start()

	<-ctx.Done()
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	api.Shutdown(ctx)
	srv.Shutdown(ctx)

	os.Exit(exitSuccess)
}
