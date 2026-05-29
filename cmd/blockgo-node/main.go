package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blockgo/internal/api"
	"blockgo/internal/config"
	"blockgo/internal/node"
	"blockgo/internal/version"
)

const (
	httpReadHeaderTimeout = 5 * time.Second
	httpReadTimeout       = 10 * time.Second
	httpWriteTimeout      = 10 * time.Second
	httpIdleTimeout       = 60 * time.Second
)

func main() {
	var (
		configPath = flag.String("config", "", "path to node config file")
		showVer    = flag.Bool("version", false, "print version information")
	)
	flag.Parse()

	if *showVer {
		fmt.Println(version.String())
		return
	}

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "error: -config is required")
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.LoadNodeConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load node config: %v\n", err)
		os.Exit(1)
	}

	genesis, err := config.LoadGenesis(cfg.GenesisFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load genesis: %v\n", err)
		os.Exit(1)
	}

	n, err := node.New(cfg.NodeConfig, genesis, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create node: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = n.Stop() }()

	if err := n.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start node: %v\n", err)
		os.Exit(1)
	}

	var httpServer *http.Server
	if cfg.HTTPAddr != "" {
		apiServer := api.NewServer(logger, n)
		httpServer = &http.Server{
			Addr:              cfg.HTTPAddr,
			Handler:           apiServer.Handler(),
			ReadHeaderTimeout: httpReadHeaderTimeout,
			ReadTimeout:       httpReadTimeout,
			WriteTimeout:      httpWriteTimeout,
			IdleTimeout:       httpIdleTimeout,
		}

		go func() {
			logger.Info("http api listening", "addr", cfg.HTTPAddr)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("http api failed", "error", err)
			}
		}()
	}

	logger.Info("node started", "node_id", cfg.NodeID, "listen_addr", cfg.ListenAddr)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	if httpServer != nil {
		_ = httpServer.Shutdown(context.Background())
	}
}
