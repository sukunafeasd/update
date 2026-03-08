package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"universald/internal/api"
	"universald/internal/config"
	"universald/internal/db"
	"universald/internal/panel"
)

const version = "1.4.4"

func main() {
	cfg := config.Load()

	bind := flag.String("bind", cfg.BindAddress, "HTTP bind address")
	dbPath := flag.String("db", cfg.DBPath, "SQLite database path")
	uploadsPath := flag.String("uploads", cfg.UploadsDir, "Directory for Painel Dief uploads")
	webPath := flag.String("web", cfg.StaticDir, "Static web directory")
	safeMode := flag.Bool("safe-mode", cfg.SafeMode, "Enable anti-cheat safe mode")
	openBrowser := flag.Bool("open", cfg.OpenBrowser, "Open dashboard in browser on startup")
	flag.Parse()

	if err := ensureRuntimePaths(*dbPath, *uploadsPath); err != nil {
		log.Fatalf("prepare runtime paths: %v", err)
	}

	store, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close db: %v", err)
		}
	}()

	panelSvc := panel.NewService(store, *uploadsPath)
	if err := panelSvc.EnsureBootstrapped(); err != nil {
		log.Fatalf("bootstrap painel dief: %v", err)
	}
	runtimeCtx, runtimeCancel := context.WithCancel(context.Background())
	defer runtimeCancel()
	panelSvc.StartMaintenance(runtimeCtx, cfg.MaintenanceInterval)

	server := api.NewPanelServer(*webPath, version, *safeMode, panelSvc, api.ServerOptions{
		AppEnv:           cfg.AppEnv,
		PublicOrigin:     cfg.PublicOrigin,
		DBPath:           *dbPath,
		OpsToken:         cfg.OpsToken,
		DownloadPassword: cfg.DownloadPassword,
	})
	httpServer := &http.Server{
		Addr:              *bind,
		Handler:           server.Handler(),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	go func() {
		log.Printf("Painel Dief running at http://%s", *bind)
		log.Printf("app env: %s", cfg.AppEnv)
		if strings.TrimSpace(cfg.PublicOrigin) != "" {
			log.Printf("public origin: %s", cfg.PublicOrigin)
		}
		log.Printf("safe mode: %v", *safeMode)
		if *openBrowser {
			go openDashboard(*bind)
		}
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("shutdown requested")
	runtimeCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func ensureRuntimePaths(dbPath, uploadsPath string) error {
	dbPath = strings.TrimSpace(dbPath)
	uploadsPath = strings.TrimSpace(uploadsPath)

	if dbPath != "" {
		dbDir := filepath.Dir(dbPath)
		if dbDir != "" && dbDir != "." {
			if err := os.MkdirAll(dbDir, 0o755); err != nil {
				return err
			}
		}
	}
	if uploadsPath != "" {
		if err := os.MkdirAll(uploadsPath, 0o755); err != nil {
			return err
		}
	}
	if dbPath == "" && uploadsPath == "" {
		return errors.New("runtime paths vazios")
	}
	return nil
}

func openDashboard(bind string) {
	time.Sleep(900 * time.Millisecond)
	url := "http://" + bind

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("open browser failed: %v", err)
	}
}
