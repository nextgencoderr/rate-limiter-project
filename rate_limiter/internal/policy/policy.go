package policy

import (
	"sync"
)

type Algorithm string

const (
	TokenBucket   Algorithm = "token_bucket"
	SlidingWindow Algorithm = "sliding_window"
)

type Policy struct {
	ClientID   string    `json:"client_id"`
	Algorithm  Algorithm `json:"algorithm"`
	Limit      int       `json:"limit"`       // max requests (sliding) or bucket capacity (token)
	WindowSec  int       `json:"window_sec"`  // sliding window duration
	RefillRate float64   `json:"refill_rate"` // tokens per second (token bucket)
}

type Store struct {
	mu       sync.RWMutex
	policies map[string]Policy
	defaults Policy
}

func NewStore(defaults Policy) *Store {
	return &Store{
		policies: make(map[string]Policy),
		defaults: defaults,
	}
}

func (s *Store) Get(clientID string) Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.policies[clientID]; ok {
		return p
	}
	p := s.defaults
	p.ClientID = clientID
	return p
}

func (s *Store) Set(p Policy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[p.ClientID] = p
}

func (s *Store) List() []Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Policy, 0, len(s.policies))
	for _, p := range s.policies {
		out = append(out, p)
	}
	return out
}

func (s *Store) Delete(clientID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.policies[clientID]; !ok {
		return false
	}
	delete(s.policies, clientID)
	return true
}
