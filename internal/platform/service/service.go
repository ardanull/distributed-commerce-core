package service

import (
    "context"
    "log/slog"
    "net/http"
    "os/signal"
    "syscall"
    "time"

    "github.com/arda/distributed-commerce-core/internal/platform/config"
)

type Runnable interface {
    Routes() http.Handler
    Background(context.Context) error
}

func Run(cfg config.Config, logger *slog.Logger, app Runnable) error {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{
        Addr:    cfg.HTTPAddr,
        Handler: app.Routes(),
    }

    go func() {
        logger.Info("http_listen", "addr", cfg.HTTPAddr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Error("http_server_failed", "error", err)
            stop()
        }
    }()

    go func() {
        if err := app.Background(ctx); err != nil {
            logger.Error("background_failed", "error", err)
            stop()
        }
    }()

    <-ctx.Done()

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return srv.Shutdown(shutdownCtx)
}
