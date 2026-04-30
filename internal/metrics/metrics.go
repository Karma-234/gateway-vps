package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	TransactionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iso8583_transactions_total",
			Help: "Total number of processed ISO 8583 transactions",
		},
		[]string{"mti", "response_code", "merchant_id"},
	)

	TransactionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "iso8583_transaction_duration_seconds",
			Help:    "Duration of ISO 8583 transaction processing",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"mti"},
	)

	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "iso8583_active_connections",
			Help: "Current number of active ISO 8583 connections",
		},
	)
)

func init() {
	prometheus.MustRegister(TransactionsTotal)
	prometheus.MustRegister(TransactionDuration)
	prometheus.MustRegister(ActiveConnections)
}

// Handler returns Prometheus metrics handler for /metrics endpoint
func Handler() http.Handler {
	return promhttp.Handler()
}
