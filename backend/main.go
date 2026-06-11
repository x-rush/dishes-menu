package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"dishes-menu/internal/api"
	"dishes-menu/internal/dao"
	"dishes-menu/internal/db"
)

//go:embed all:migrations
var migrationsFS embed.FS

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

	if err := runMigrations(mysqlDB); err != nil {
		logger.Error("run migrations", "err", err)
		os.Exit(1)
	}
	logger.Info("migrations up-to-date")

	if n, err := seedDishes(ctx, mysqlDB); err != nil {
		logger.Error("seed dishes", "err", err)
		os.Exit(1)
	} else if n > 0 {
		logger.Info("seeded default dishes", "count", n)
	} else {
		logger.Info("dishes table already populated, skip seed")
	}

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

func runMigrations(mysqlDB *sqlx.DB) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("iofs: %w", err)
	}
	driver, err := mysqlmigrate.WithInstance(mysqlDB.DB, &mysqlmigrate.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func seedDishes(ctx context.Context, mysqlDB *sqlx.DB) (int, error) {
	repo := dao.NewDishRepo(mysqlDB)
	return repo.SeedDefaultDishes(ctx)
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}
