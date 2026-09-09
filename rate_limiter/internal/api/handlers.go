package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ratelimiter/service/internal/limiter"
	"github.com/ratelimiter/service/internal/metrics"
	"github.com/ratelimiter/service/internal/policy"
)

type Handler struct {
	limiter *limiter.Service
	store   *policy.Store
}

func New(l *limiter.Service, store *policy.Store) *Handler {
	return &Handler{limiter: l, store: store}
}

type checkRequest struct {
	ClientID string `json:"client_id"`
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req checkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ClientID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "client_id required"})
		return
	}

	p := h.store.Get(req.ClientID)
	res, err := h.limiter.Allow(r.Context(), p)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "rate limiter unavailable"})
		return
	}

	result := "allowed"
	if !res.Allowed {
		result = "throttled"
		metrics.ThrottleEvents.WithLabelValues(req.ClientID, res.Algorithm).Inc()
	}
	metrics.RequestsTotal.WithLabelValues(req.ClientID, res.Algorithm, result).Inc()

	status := http.StatusOK
	if !res.Allowed {
		status = http.StatusTooManyRequests
	}
	writeJSON(w, status, res)
}

func (h *Handler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, h.store.List())
}

func (h *Handler) Policy(w http.ResponseWriter, r *http.Request) {
	clientID := strings.TrimPrefix(r.URL.Path, "/v1/policies/")
	if clientID == "" || strings.Contains(clientID, "/") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid client_id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.store.Get(clientID))
	case http.MethodPut:
		var p policy.Policy
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
			return
		}
		if err := validatePolicy(p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		p.ClientID = clientID
		h.store.Set(p)
		metrics.PolicyUpdates.Inc()
		writeJSON(w, http.StatusOK, p)
	case http.MethodDelete:
		if !h.store.Delete(clientID) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "policy not found"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(w)
	}
}

func (h *Handler) Live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.limiter.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "redis unavailable"})
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

func validatePolicy(p policy.Policy) error {
	switch p.Algorithm {
	case policy.TokenBucket, policy.SlidingWindow:
	default:
		return errors.New("algorithm must be token_bucket or sliding_window")
	}
	if p.Limit <= 0 {
		return errors.New("limit must be positive")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}
