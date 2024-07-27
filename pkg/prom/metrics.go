package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var TotalRequestsStartedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "connect_total_requests_started",
	Help: "Total number of requests started by a ConnectRPC service",
}, []string{"stream_type", "service", "method"})

var TotalRequestsCompletedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "connect_total_requests_completed",
	Help: "Total number of requests completed by a ConnectRPC service",
}, []string{"stream_type", "service", "method", "code"})

var TotalRequestBytesCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "connect_total_request_bytes",
	Help: "Total number of request bytes sent by a ConnectRPC service",
}, []string{"stream_type", "service", "method"})

var TotalResponseBytesCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "connect_total_response_bytes",
	Help: "Total number of response bytes received by a ConnectRPC service",
}, []string{"stream_type", "service", "method"})

var RequestLatencySecondsHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "connect_request_latency_seconds",
	Help:    "Request latency in seconds for a ConnectRPC service",
	Buckets: prometheus.ExponentialBuckets(0.0001, 2, 20),
}, []string{"stream_type", "service", "method", "code"})

func RegisterMetrics(registry prometheus.Registerer) {
	registry.MustRegister(TotalRequestsStartedCounter)
	registry.MustRegister(TotalRequestsCompletedCounter)
	registry.MustRegister(RequestLatencySecondsHistogram)
}
