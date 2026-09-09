package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.Live)
	mux.HandleFunc("/readyz", h.Ready)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/v1/check", h.Check)
	mux.HandleFunc("/v1/policies", h.ListPolicies)
	mux.HandleFunc("/v1/policies/", h.Policy)
	return mux
}
