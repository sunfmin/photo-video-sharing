package testutil

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

// MockStorage is an in-memory storage implementation for testing
type MockStorage struct {
	mu    sync.RWMutex
	files map[string][]byte // key -> file content
}

// NewMockStorage creates a new mock storage instance
func NewMockStorage() *MockStorage {
	return &MockStorage{
		files: make(map[string][]byte),
	}
}

// Upload stores file content in memory
func (m *MockStorage) Upload(key string, reader io.Reader, size int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}

	m.files[key] = content
	return nil
}

// Download retrieves file content from memory
func (m *MockStorage) Download(key string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	content, exists := m.files[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}

	return io.NopCloser(bytes.NewReader(content)), nil
}

// Delete removes file from memory
func (m *MockStorage) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.files, key)
	return nil
}

// GetPresignedURL returns a mock URL (not actually presigned)
func (m *MockStorage) GetPresignedURL(key string, expiry time.Duration) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.files[key]; !exists {
		return "", fmt.Errorf("file not found: %s", key)
	}

	return fmt.Sprintf("http://mock-storage/download/%s", key), nil
}

// Exists checks if a file exists in mock storage
func (m *MockStorage) Exists(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.files[key]
	return exists
}

// Clear removes all files from mock storage
func (m *MockStorage) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.files = make(map[string][]byte)
}

// Count returns the number of files in mock storage
func (m *MockStorage) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.files)
}

