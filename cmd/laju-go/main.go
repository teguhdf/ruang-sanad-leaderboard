package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/maulanashalihin/laju-go/app/cache"
	"github.com/maulanashalihin/laju-go/app/config"
	"github.com/maulanashalihin/laju-go/app/handlers"
	"github.com/maulanashalihin/laju-go/app/middlewares"
	"github.com/maulanashalihin/laju-go/app/queries"
	"github.com/maulanashalihin/laju-go/app/services"
	"github.com/maulanashalihin/laju-go/app/session"
	"github.com/maulanashalihin/laju-go/routes"
	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

var (
	Version = "v1.1.0"
	Commit  = "none"
)

func main() {
	showVersion := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("laju-go %s (commit: %s)\n", Version, Commit)
		os.Exit(0)
	}

	cfg := config.Load()

	logLevel := slog.LevelInfo
	if cfg.AppEnv == "development" {
		logLevel = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: logLevel}
	var handler slog.Handler
	if cfg.AppEnv == "development" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
	slog.Info("starting", "version", Version, "commit", Commit, "env", cfg.AppEnv)

	db, err := initDatabase(cfg.DBPath)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(db, "./migrations"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	querier := queries.NewQuerier(db)
	sessionCache := cache.NewSessionCache()
	sessionStore := session.New(querier, sessionCache, cfg.SessionTTL)
	sessionStore.SetSecure(cfg.AppEnv == "production")
	startBackgroundCleanup(querier, sessionCache)

	authService := services.NewAuthService(querier, services.AuthServiceConfig{
		SessionSecret:      cfg.SessionSecret,
		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
		GoogleRedirectURL:  cfg.GoogleRedirectURL,
	})
	userService := services.NewUserService(querier)
	memberService := services.NewMemberService(querier)
	fingerprintImportService := services.NewFingerprintImportService(querier)
	attendanceService := services.NewAttendanceService(querier)

	assetService := services.NewAssetService("./dist/.vite/manifest.json", ".vite-port", cfg.IsDevelopment())
	inertiaService := services.NewInertiaService(assetService, sessionStore)

	if err := os.MkdirAll("storage/uploads", 0755); err != nil {
		slog.Error("failed to create uploads directory", "error", err)
		os.Exit(1)
	}
	if err := os.MkdirAll("storage/avatars", 0755); err != nil {
		slog.Error("failed to create avatars directory", "error", err)
		os.Exit(1)
	}
	csrfMiddleware := routes.SetupCSRFMiddleware(cfg.SessionSecret, cfg.AppEnv == "production")

	appURL := routes.GetAppURL(cfg.AppPort, cfg.AppEnv)
	mailerService := services.NewMailerService(querier, services.MailConfig{
		Driver:          services.MailDriver(cfg.MailDriver),
		From:            cfg.MailFrom,
		FromName:        cfg.MailFromName,
		ResendAPIKey:    cfg.ResendAPIKey,
		MailtrapToken:   cfg.MailtrapAPIToken,
		MailtrapInboxID: cfg.MailtrapInboxID,
	}, appURL)

	uploadHandler := handlers.NewUploadHandler(sessionStore, userService, "storage/uploads")
	routeHandlers := routes.Handlers{
		Public: handlers.NewPublicHandler(authService, userService, inertiaService, assetService),
		Auth:   handlers.NewAuthHandler(authService, sessionStore, inertiaService, mailerService),
		App:    handlers.NewAppHandler(userService, memberService, fingerprintImportService, attendanceService, sessionStore, inertiaService),
		Upload: uploadHandler,
	}

	passwordResetHandler := routes.SetupPasswordResetHandler(
		mailerService,
		userService,
		sessionStore,
		inertiaService,
	)
	routeHandlers.PasswordReset = passwordResetHandler
	app := fiber.New(fiber.Config{
		AppName:      "Ruang Sanad Ops",
		ErrorHandler: customErrorHandler,
		StreamRequestBody: true,
		ReadBufferSize: 64 * 1024,
	})

	if cfg.IsDevelopment() {
		app.Use(logger.New())
	}
	app.Use(recover.New())
	app.Use(middlewares.SecurityHeaders())
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	app.Use(inertiaService.Middleware())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.AllowedOrigins, ","),
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Inertia, X-Inertia-Version, X-Requested-With, Tus-Resumable, Upload-Length, Upload-Offset, Upload-Metadata, Upload-Draft-Interop-Version, Upload-Complete, Upload-Incomplete, Upload-Concat, Upload-Defer-Length",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": Version,
		})
	})

	routes.SetupRoutes(app, routeHandlers, sessionStore, userService, mailerService, csrfMiddleware)

	go func() {
		slog.Info("server listening", "port", cfg.AppPort)
		if err := app.Listen(":" + cfg.AppPort); err != nil {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutting down", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		slog.Error("shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func initDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA cache_size = -16000"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA mmap_size = 268435456"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA temp_store = MEMORY"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA wal_autocheckpoint = 1000"); err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	logDatabaseOptimizations(db)
	return db, nil
}

func logDatabaseOptimizations(db *sql.DB) {
	var journalMode, synchronous string
	var cacheSize, busyTimeout, mmapSize, walAutocheckpoint int

	err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		slog.Warn("could not verify journal_mode", "error", err)
	}
	err = db.QueryRow("PRAGMA synchronous").Scan(&synchronous)
	if err != nil {
		slog.Warn("could not verify synchronous", "error", err)
	}
	err = db.QueryRow("PRAGMA cache_size").Scan(&cacheSize)
	if err != nil {
		slog.Warn("could not verify cache_size", "error", err)
	}
	err = db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout)
	if err != nil {
		slog.Warn("could not verify busy_timeout", "error", err)
	}
	err = db.QueryRow("PRAGMA mmap_size").Scan(&mmapSize)
	if err != nil {
		slog.Warn("could not verify mmap_size", "error", err)
	}
	err = db.QueryRow("PRAGMA wal_autocheckpoint").Scan(&walAutocheckpoint)
	if err != nil {
		slog.Warn("could not verify wal_autocheckpoint", "error", err)
	}

	slog.Info("sqlite optimizations",
		"journal_mode", journalMode,
		"synchronous", synchronous,
		"cache_size_kb", cacheSize,
		"mmap_size_kb", mmapSize,
		"wal_autocheckpoint", walAutocheckpoint,
		"busy_timeout_ms", busyTimeout,
	)
}

func startBackgroundCleanup(querier *queries.Querier, sessionCache *cache.SessionCache) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		if err := querier.DeleteExpiredSessions(context.Background()); err != nil {
			slog.Error("cleanup: failed to delete expired sessions", "error", err)
		} else {
			slog.Debug("cleanup: expired sessions deleted")
		}
		if err := querier.DeleteExpiredPasswordResets(context.Background()); err != nil {
			slog.Error("cleanup: failed to delete expired password resets", "error", err)
		} else {
			slog.Debug("cleanup: expired password resets deleted")
		}
		sessionCache.EvictExpired()
		if err := querier.DeleteExpiredEmailVerifications(context.Background()); err != nil {
			slog.Error("cleanup: failed to delete expired email verifications", "error", err)
		} else {
			slog.Debug("cleanup: expired email verifications deleted")
		}

		for range ticker.C {
			if err := querier.DeleteExpiredSessions(context.Background()); err != nil {
				slog.Error("cleanup: failed to delete expired sessions", "error", err)
			}
			if err := querier.DeleteExpiredPasswordResets(context.Background()); err != nil {
				slog.Error("cleanup: failed to delete expired password resets", "error", err)
			}
			sessionCache.EvictExpired()
			if err := querier.DeleteExpiredEmailVerifications(context.Background()); err != nil {
				slog.Error("cleanup: failed to delete expired email verifications", "error", err)
			}
		}
	}()
	slog.Info("background cleanup started (interval: 1h)")
}

func runMigrations(db *sql.DB, migrationsDir string) error {
	goose.SetBaseFS(nil)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	if err := goose.Up(db, migrationsDir); err != nil {
		return err
	}
	return nil
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	slog.Error("request error", "method", c.Method(), "path", c.Path(), "status", code, "error", err)

	msg := "Internal Server Error"
	if code < 500 {
		msg = err.Error()
	}

	return c.Status(code).JSON(fiber.Map{
		"error": msg,
	})
}
