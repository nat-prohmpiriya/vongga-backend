package utils

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Request metrics
	HttpRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Request Duration
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.1, 0.3, 0.5, 0.7, 1, 3, 5, 7, 10},
		},
		[]string{"method", "endpoint"},
	)

	// Error Counter
	ErrorTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "error_total",
			Help: "Total number of errors",
		},
		[]string{"type"},
	)

	// Business metrics ตามความต้องการ เช่น
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users",
		},
	)
)
