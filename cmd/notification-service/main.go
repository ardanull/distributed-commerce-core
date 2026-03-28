package main

import (
    "context"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/prometheus/client_golang/prometheus/promhttp"

    "github.com/ardanull/distributed-commerce-core/internal/platform/config"
    "github.com/ardanull/distributed-commerce-core/internal/platform/httpx"
    "github.com/ardanull/distributed-commerce-core/internal/platform/logging"
    "github.com/ardanull/distributed-commerce-core/internal/platform/natsx"
    "github.com/ardanull/distributed-commerce-core/internal/platform/postgres"
    "github.com/ardanull/distributed-commerce-core/internal/platform/redisx"
    "github.com/ardanull/distributed-commerce-core/internal/platform/service"
    "github.com/ardanull/distributed-commerce-core/internal/platform/telemetry"
    "github.com/ardanull/distributed-commerce-core/internal/platform/inbox"
    "github.com/ardanull/distributed-commerce-core/internal/services/notification"
)

type app struct {
    router chi.Router
    bg     func(context.Context) error
}

func (a app) Routes() http.Handler { return a.router }
func (a app) Background(ctx context.Context) error { return a.bg(ctx) }

func main() {
    cfg := config.Load()
    logger := logging.New(cfg.AppName)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    shutdownTelemetry, err := telemetry.Init(ctx, cfg.OTLPEndpoint, cfg.AppName, cfg.TraceEnabled)
    if err != nil {
        panic(err)
    }
    defer shutdownTelemetry(context.Background())

    db, err := postgres.Connect(ctx, cfg.PostgresDSN)
    if err != nil {
        panic(err)
    }

    redisClient := redisx.New(cfg.RedisAddr)
    if err := redisx.Ping(ctx, redisClient); err != nil {
        panic(err)
    }

    bus, err := natsx.Connect(cfg.NATSURL)
    if err != nil {
        panic(err)
    }
    if err := bus.EnsureStream(); err != nil {
        panic(err)
    }

    router := httpx.NewRouter(logger)
    router.Handle("/metrics", promhttp.Handler())


    svc := notification.NewService(db, inbox.New(db), logger)
    bg := func(ctx context.Context) error {
        _, err := bus.SubscribeDurable("notification.send.*", "notification-send", svc.HandleSend)
        return err
    }


    if err := service.Run(cfg, logger, app{router: router, bg: bg}); err != nil {
        panic(err)
    }
}
