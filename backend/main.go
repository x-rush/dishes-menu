package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"dishes-menu/internal/api"
	"dishes-menu/internal/db"
)

//go:embed all:web/dist
var webDistFS embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Try .env in CWD, then project root (../.env when running from backend/).
	// In Docker, env_file injects vars so this is a no-op there.
	for _, p := range []string{".env", "../.env", "/app/.env"} {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
			break
		}
	}

	port := getenv("PORT", "8080")
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		logger.Error("MYSQL_DSN is required")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mysqlDB, err := db.Open(ctx, dsn)
	if err != nil {
		logger.Error("connect mysql", "err", err)
		os.Exit(1)
	}
	defer mysqlDB.Close()
	logger.Info("connected to mysql")

	if getenv("GIN_MODE", "") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	h := api.NewHandlers(mysqlDB)

	subFS, err := fs.Sub(webDistFS, "web/dist")
	if err != nil {
		logger.Error("sub web/dist", "err", err)
		os.Exit(1)
	}
	if _, err := fs.Stat(subFS, "index.html"); err != nil {
		logger.Warn("frontend not built; API-only mode", "hint", "run `npm run build` inside frontend/")
		subFS = nil
	}
	api.RegisterRoutes(r, h, subFS)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		logger.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "err", err)
	}
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}
