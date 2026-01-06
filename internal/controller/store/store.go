package store

import (
	"sync"
	"time"
)

type ServerStatus struct {
	LastSeen    time.Time
	Load        float64
	IsOnline    bool
	Policy      string // "sleep" or "stay_awake"
	Override    bool   // If true, User manually set policy to "stay_awake"
	ForceSleep  bool   // If true, User manually requested "sleep" (one-shot)
}

type Store struct {
	mu     sync.RWMutex
	status ServerStatus
}

func NewStore() *Store {
	return &Store{
		status: ServerStatus{
			Policy: "sleep", // Default policy
		},
	}
}

func (s *Store) UpdateHeartbeat(load float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.status.LastSeen = time.Now()
	s.status.Load = load
	s.status.IsOnline = true
	
	// Reset ForceSleep if it was active and we're seeing a heartbeat?
	// Actually, the heartbeat will Consume it.
}

func (s *Store) RequestShutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.ForceSleep = true
}

func (s *Store) ConsumeShutdownRequest() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status.ForceSleep {
		s.status.ForceSleep = false
		return true
	}
	return false
}

func (s *Store) GetStatus() ServerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Check if timeout (offline detection)
	if time.Since(s.status.LastSeen) > 2*time.Minute {
		s.status.IsOnline = false
	}
	
	return s.status
}

func (s *Store) SetOverride(keepAwake bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Override = keepAwake
}
