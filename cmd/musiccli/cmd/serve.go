package cmd

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"choinhaccli/internal/config"
	"choinhaccli/internal/library"
	"choinhaccli/internal/web"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the musicweb HTTP server",
	RunE:  runServe,
}

var serveFlags struct {
	musicDir string
	port     int
	bind     string
	dbPath   string
	scan     bool
}

func init() {
	serveCmd.Flags().StringVar(&serveFlags.musicDir, "music-dir", "", "path to music library (overrides config)")
	serveCmd.Flags().IntVar(&serveFlags.port, "port", 0, "HTTP port (default 8080)")
	serveCmd.Flags().StringVar(&serveFlags.bind, "bind", "", "bind address (default 127.0.0.1)")
	serveCmd.Flags().StringVar(&serveFlags.dbPath, "db", "", "SQLite DB path (overrides config)")
	serveCmd.Flags().BoolVar(&serveFlags.scan, "scan", true, "scan music dir on startup")
}

func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("config load warning: %v (using defaults)", err)
	}

	if serveFlags.musicDir != "" {
		cfg.MusicDir = serveFlags.musicDir
	}
	if serveFlags.port != 0 {
		cfg.Port = serveFlags.port
	}
	if serveFlags.bind != "" {
		cfg.BindAddr = serveFlags.bind
	}
	if serveFlags.dbPath != "" {
		cfg.DBPath = serveFlags.dbPath
	}

	dataDir := filepath.Dir(cfg.DBPath)
	coverDir := filepath.Join(dataDir, "covers")

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	store, err := library.OpenStore(cfg.DBPath)
	if err != nil {
		return err
	}
	defer store.Close()

	// Background scan on startup
	if serveFlags.scan && cfg.MusicDir != "" {
		go func() {
			log.Printf("scanning %s ...", cfg.MusicDir)
			for p := range library.ScanAsync(cfg.MusicDir, store, coverDir) {
				if p.Done {
					log.Printf("scan complete: %d tracks", p.Scanned)
				}
			}
		}()
	}

	srv := web.NewServer(cfg.BindAddr, cfg.Port, store, cfg.MusicDir, coverDir)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
