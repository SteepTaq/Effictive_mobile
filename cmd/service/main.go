package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rest-service/internal/config"
	"rest-service/internal/handlers"
	"rest-service/internal/repository"
	"rest-service/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log := logger.Setup(cfg.Logger.Level)
	slog.SetDefault(log)

	slog.Info("starting subscription service")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("database connection open")

	repo := repository.NewRepository(pool)
	handler := handlers.NewHandler(repo, log)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	router.Get("/api-docs/*", httpSwagger.Handler(httpSwagger.URL("/static/swagger.json")))
	router.Handle("/*", http.FileServer(http.Dir("static")))
	router.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", handler.CreateSubscription)
		r.Get("/", handler.ListSubscriptions)
		r.Get("/summary", handler.SummarySubscriptions)
		r.Get("/{id}", handler.GetSubscription)
		r.Put("/{id}", handler.UpdateSubscription)
		r.Delete("/{id}", handler.DeleteSubscription)
	})

	server := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("http server started", "port", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start http server", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("get signal, stop service")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown http server", "error", err)
	}
	slog.Info("service stopped")
}
