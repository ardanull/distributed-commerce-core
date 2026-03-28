package config

import (
    "fmt"
    "os"
)

type Config struct {
    AppName             string
    HTTPAddr            string
    PostgresDSN         string
    NATSURL             string
    RedisAddr           string
    TraceEnabled        bool
    OTLPEndpoint        string
}

func Load() Config {
    cfg := Config{
        AppName:     getenv("APP_NAME", "service"),
        HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
        PostgresDSN: getenv("POSTGRES_DSN", "postgres://app:app@localhost:5432/commerce?sslmode=disable"),
        NATSURL:     getenv("NATS_URL", "nats://localhost:4222"),
        RedisAddr:   getenv("REDIS_ADDR", "localhost:6379"),
        OTLPEndpoint: getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
        TraceEnabled: getenv("TRACE_ENABLED", "false") == "true",
    }
    if cfg.PostgresDSN == "" || cfg.NATSURL == "" {
        panic("missing critical configuration")
    }
    return cfg
}

func (c Config) Validate() error {
    if c.AppName == "" {
        return fmt.Errorf("APP_NAME is required")
    }
    if c.HTTPAddr == "" {
        return fmt.Errorf("HTTP_ADDR is required")
    }
    return nil
}

func getenv(k, d string) string {
    v := os.Getenv(k)
    if v == "" {
        return d
    }
    return v
}
