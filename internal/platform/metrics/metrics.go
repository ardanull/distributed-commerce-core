package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
    HTTPRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests.",
    }, []string{"service", "path", "method", "status"})

    OutboxPublishes = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "outbox_publish_total",
        Help: "Published outbox events.",
    }, []string{"service", "subject"})

    ConsumerDedupe = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "consumer_dedupe_total",
        Help: "Dedup hits per consumer.",
    }, []string{"service", "consumer"})
)

func Register() {
    prometheus.MustRegister(HTTPRequests, OutboxPublishes, ConsumerDedupe)
}
