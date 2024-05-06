package store

import (
	"fmt"
	"sync"

	"github.com/AR1011/ssh-webhook/types"
)

type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]types.WebhookConfig
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]types.WebhookConfig),
	}
}

func (m *MemoryStore) GetByID(id string) (types.WebhookConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[id]
	if !ok {
		return types.WebhookConfig{}, fmt.Errorf("session not found")
	}

	return session, nil
}

func (m *MemoryStore) GetBySessionID(id string) (types.WebhookConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, session := range m.sessions {
		if session.ActiveSession.SessionID == id {
			return session, nil
		}
	}

	return types.WebhookConfig{}, fmt.Errorf("session not found")
}

func (m *MemoryStore) Set(id string, value types.WebhookConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[id] = value
	return nil
}

func (m *MemoryStore) Update(id string, value types.WebhookConfig) error {
	return m.Set(id, value)
}

func (m *MemoryStore) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, id)
	return nil
}

func (m *MemoryStore) IsPortAvailable(port int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, session := range m.sessions {
		if session.InternalServerSocket.Port == int64(port) {
			return false
		}
	}

	return true
}

func (m *MemoryStore) FindByPublicKeyAndSocket(publicKey string, host string, port int) (types.WebhookConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, session := range m.sessions {
		if session.PubKey == publicKey && session.InternalServerSocket.Host == host && session.InternalServerSocket.Port == int64(port) {
			return session, nil
		}
	}

	return types.WebhookConfig{}, fmt.Errorf("session not found")
}

var _ Store = &MemoryStore{}
