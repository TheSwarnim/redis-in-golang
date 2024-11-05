package store

import (
	"sync"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type Store struct {
    mu   sync.RWMutex
    data map[string]resp.Value
}

func NewStore() *Store {
    return &Store{
        data: make(map[string]resp.Value),
    }
}

func (s *Store) Get(key string) (resp.Value, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    val, exists := s.data[key]
    if !exists {
        return resp.Value{}, false
    }

    if val.IsExpired() {
        s.mu.RUnlock()
        s.mu.Lock()
        delete(s.data, key)
        s.mu.Unlock()
        s.mu.RLock()
        return resp.Value{}, false
    }

    return val, true
}

func (s *Store) Set(key string, value resp.Value) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
}
