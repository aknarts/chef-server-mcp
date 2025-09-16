package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aknarts/chef-server-mcp/internal/chefapi"
	"github.com/aknarts/chef-server-mcp/internal/config"
	"github.com/aknarts/chef-server-mcp/internal/server"
	"github.com/aknarts/chef-server-mcp/internal/version"
)

func main() {
	cfg := config.LoadFromEnv()
	log.Printf("chef-mcp starting version=%s port=%s knife_fallback=%t", version.Version, cfg.Port, cfg.KnifeFallback)

	var chefClient *chefapi.ChefAPI
	if cfg.ChefUser != "" && cfg.ChefKeyPath != "" && cfg.ChefServerURL != "" {
		c, err := chefapi.NewChefAPI(cfg.ChefUser, cfg.ChefKeyPath, cfg.ChefServerURL)
		if err != nil {
			if !cfg.KnifeFallback {
				log.Fatalf("failed to init Chef API client: %v", err)
			}
			log.Printf("warning: Chef API client init failed, relying on knife fallback: %v", err)
		} else {
			chefClient = c
		}
	} else if !cfg.KnifeFallback {
		log.Fatalf("Chef API credentials missing and knife fallback disabled")
	} else {
		log.Printf("Chef API credentials incomplete; operating with knife fallback only")
	}

	srv := server.New(cfg, chefClient)
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("signal received: %s; shutting down", sig)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Printf("shutdown complete")
}
