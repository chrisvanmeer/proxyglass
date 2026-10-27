package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chrisvanmeer/proxyglass/internal/auth"
	"github.com/chrisvanmeer/proxyglass/internal/build"
	"github.com/chrisvanmeer/proxyglass/internal/config"
	"github.com/chrisvanmeer/proxyglass/internal/installer"
	"github.com/chrisvanmeer/proxyglass/internal/logger"
	"github.com/chrisvanmeer/proxyglass/internal/logo"
	"github.com/chrisvanmeer/proxyglass/internal/server"
)

func tokenTTL(cfg *config.AppConfig) time.Duration {
	if cfg.TokenTTLDays <= 0 {
		return 0
	}
	return time.Duration(cfg.TokenTTLDays) * 24 * time.Hour
}

func main() {
	configPath := flag.String("config", "", "Path to configuration file (defaults to PROXYGLASS_CONFIG or /etc/proxyglass/config.yml)")
	doInstall := flag.Bool("install", false, "Install Proxyglass service, user, and configuration")
	doUninstall := flag.Bool("uninstall", false, "Uninstall Proxyglass service (keeps configuration and data)")
	genToken := flag.Bool("generate-token", false, "Generate an administrative Bearer token")
	revokeToken := flag.String("revoke-token", "", "Revoke a Bearer token")
	quiet := flag.Bool("quiet", false, "Output raw token only")
	flag.BoolVar(quiet, "q", false, "Output raw token only (shorthand)")
	showVersion := flag.Bool("version", false, "Print build version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("proxyglass %s\n", build.Version)
		os.Exit(0)
	}

	if *doInstall {
		if err := installer.RunInstall(); err != nil {
			logger.Error("Installation failed: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *doUninstall {
		if err := installer.RunUninstall(); err != nil {
			logger.Error("Uninstallation failed: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		// try local config
		cfg, err = config.Load("config.yaml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}
	}

	if *genToken {
		token, _, err := auth.GenerateToken(cfg.TokensFile, "", tokenTTL(cfg))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate token: %v\n", err)
			os.Exit(1)
		}
		if *quiet {
			fmt.Print(token)
		} else {
			fmt.Printf("Generated new administrative Bearer Token: %s\n", token)
		}
		os.Exit(0)
	}

	if *revokeToken != "" {
		if err := auth.RevokeToken(cfg.TokensFile, *revokeToken); err != nil {
			logger.Error("Failed to revoke token: %v", err)
			os.Exit(1)
		}
		logger.Info("Token revoked successfully")
		os.Exit(0)
	}

	logo.PrintBanner()

	if err := logger.Init(cfg.LogPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize log file %s: %v\n", cfg.LogPath, err)
	}
	defer logger.Close()

	srv, err := server.NewServer(cfg)
	if err != nil {
		logger.Error("Failed to initialize server: %v", err)
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	reload := make(chan os.Signal, 1)
	signal.Notify(reload, syscall.SIGHUP)

	start := func(instance *server.Server) {
		go func(current *server.Server) {
			if err := current.Start(); err != nil {
				logger.Error("Server error: %v", err)
				os.Exit(1)
			}
		}(instance)
	}
	start(srv)

	for {
		select {
		case <-reload:
			logger.Info("SIGHUP received: reloading configuration from %s", *configPath)
			next, err := server.NewServerFromPath(*configPath)
			if err != nil {
				logger.Error("Configuration reload failed: %v", err)
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			if err := srv.Shutdown(ctx); err != nil {
				logger.Error("Error shutting down previous server: %v", err)
			}
			cancel()
			srv = next
			start(srv)
		case <-stop:
			logger.Info("Shutting down Proxyglass...")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				logger.Error("Error shutting down server: %v", err)
			}
			logger.Info("Proxyglass stopped gracefully")
			return
		}
	}
}
