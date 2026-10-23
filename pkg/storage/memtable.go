package storage

import (
	"sync"
)

// MemTable represents an in-memory index holding active key-value pairs
type MemTable struct {
	mu sync.RWMutex
	data map[string][]byte
}

// This function initializes and returns a new empty MemTable instance
func NewMemTable() *MemTable {
	return &MemTable{
		data: make(map[string][]byte),
	}
}

// This function inserts or updates a key-value pair in memory
func (m *MemTable) Set(key, value []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Convert byte slice key to string so it can be used a Go map key
	// Make a copy of value to ensure memory safety
	valCopy := make([]byte, len(value))
	copy(valCopy, value)

	m.data[string(key)] = valCopy
}

// This functions retrieves a value by it's key from memory
// Returns the value and a boolean indicating if it exists
func (m *MemTable) Get(key []byte) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, exists := m.data[string(key)]
	if !exists {
		return nil, false
	}

	// Return a copy of the value slice to prevent external callers from mutating internal state
	valCopy := make([]byte, len(val))
	copy(valCopy, val)

	return valCopy, true
}

// This function populates the MemTable using records recovered from the WAL on startup
func (m *MemTable) LoadFromWAL(records []*Record) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, rec := range records {
		m.data[string(rec.Key)] = rec.Value
	}
}