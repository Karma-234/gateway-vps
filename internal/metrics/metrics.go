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
			Help: "Total number of ISO 8583 transactions",
		},
		[]string{"mti", "response_code"},
	)

	TransactionDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "iso8583_transaction_duration_seconds",
			Help:    "Transaction processing duration",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func init() {
	prometheus.MustRegister(TransactionsTotal)
	prometheus.MustRegister(TransactionDuration)
}

func Handler() http.Handler {
	return promhttp.Handler()
}
