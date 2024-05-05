package store

type MemoryStore struct {
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (m *MemoryStore) GetByID(id string) (string, error) {
	return "", nil
}

func (m *MemoryStore) Set(id, value string) error {
	return nil
}

var _ Store = &MemoryStore{}
