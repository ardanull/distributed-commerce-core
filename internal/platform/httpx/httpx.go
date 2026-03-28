package httpx

import (
    "context"
    "encoding/json"
    "log/slog"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
)

type statusRecorder struct {
    http.ResponseWriter
    status int
}

func (r *statusRecorder) WriteHeader(code int) {
    r.status = code
    r.ResponseWriter.WriteHeader(code)
}

func NewRouter(logger *slog.Logger) chi.Router {
    r := chi.NewRouter()
    r.Use(recovery(logger))
    r.Use(requestID())
    r.Use(loggingMiddleware(logger))
    r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); _, _ = w.Write([]byte("ok")) })
    r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); _, _ = w.Write([]byte("ok")) })
    return r
}

func JSON(w http.ResponseWriter, code int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}

func requestID() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            reqID := r.Header.Get("X-Correlation-ID")
            if reqID == "" {
                reqID = uuid.NewString()
            }
            ctx := context.WithValue(r.Context(), "correlation_id", reqID)
            w.Header().Set("X-Correlation-ID", reqID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func CorrelationID(ctx context.Context) string {
    v, _ := ctx.Value("correlation_id").(string)
    return v
}

func recovery(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    logger.Error("panic recovered", "panic", rec)
                    JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            started := time.Now()
            rec := &statusRecorder{ResponseWriter: w, status: 200}
            next.ServeHTTP(rec, r)
            logger.Info("http_request",
                "method", r.Method,
                "path", r.URL.Path,
                "status", rec.status,
                "duration_ms", time.Since(started).Milliseconds(),
                "correlation_id", CorrelationID(r.Context()),
            )
        })
    }
}
