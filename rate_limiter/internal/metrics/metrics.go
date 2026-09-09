package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_requests_total",
			Help: "Total rate-limit check requests",
		},
		[]string{"client_id", "algorithm", "result"},
	)

	ThrottleEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_throttle_events_total",
			Help: "Requests denied by the rate limiter",
		},
		[]string{"client_id", "algorithm"},
	)

	PolicyUpdates = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limit_policy_updates_total",
			Help: "Runtime policy configuration updates",
		},
	)
)
