// Package storage provides a thread-safe JSON file-based persistence layer.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileSystem abstracts file I/O for testability.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(path string) (os.FileInfo, error)
}

// OSFileSystem is the real filesystem implementation.
type OSFileSystem struct{}

func (OSFileSystem) ReadFile(path string) ([]byte, error) { return os.ReadFile(path) }
func (OSFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}
func (OSFileSystem) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (OSFileSystem) Stat(path string) (os.FileInfo, error)        { return os.Stat(path) }

// Store provides thread-safe CRUD operations on a JSON file.
type Store[T any] struct {
	mu       sync.RWMutex
	filePath string
	fs       FileSystem
}

// NewStore creates a new Store for the given entity type.
func NewStore[T any](dir, filename string, fs FileSystem) (*Store[T], error) {
	if fs == nil {
		fs = OSFileSystem{}
	}
	if err := fs.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("storage: mkdir %s: %w", dir, err)
	}
	fp := filepath.Join(dir, filename)
	if _, err := fs.Stat(fp); os.IsNotExist(err) {
		if err := fs.WriteFile(fp, []byte("[]"), 0644); err != nil {
			return nil, fmt.Errorf("storage: init %s: %w", fp, err)
		}
	}
	return &Store[T]{filePath: fp, fs: fs}, nil
}

// LoadAll reads all records from the JSON file.
func (s *Store[T]) LoadAll() ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadAllUnsafe()
}

func (s *Store[T]) loadAllUnsafe() ([]T, error) {
	data, err := s.fs.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("storage: read %s: %w", s.filePath, err)
	}
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("storage: unmarshal %s: %w", s.filePath, err)
	}
	return items, nil
}

// SaveAll writes all records to the JSON file.
func (s *Store[T]) SaveAll(items []T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveAllUnsafe(items)
}

func (s *Store[T]) saveAllUnsafe(items []T) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("storage: marshal: %w", err)
	}
	if err := s.fs.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("storage: write %s: %w", s.filePath, err)
	}
	return nil
}

// WithLock provides an exclusive lock for read-modify-write operations.
// The callback receives all items and must return the modified slice.
func (s *Store[T]) WithLock(fn func([]T) ([]T, error)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.loadAllUnsafe()
	if err != nil {
		return err
	}
	items, err = fn(items)
	if err != nil {
		return err
	}
	return s.saveAllUnsafe(items)
}

// ConfigStore handles reading and writing the config.json.
type ConfigStore struct {
	mu       sync.RWMutex
	filePath string
	fs       FileSystem
}

// NewConfigStore creates a config store.
func NewConfigStore(filePath string, fs FileSystem) *ConfigStore {
	if fs == nil {
		fs = OSFileSystem{}
	}
	return &ConfigStore{filePath: filePath, fs: fs}
}

// Load reads the config from disk.
func (c *ConfigStore) Load(cfg interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, err := c.fs.ReadFile(c.filePath)
	if err != nil {
		return fmt.Errorf("config: read: %w", err)
	}
	return json.Unmarshal(data, cfg)
}

// Save writes the config to disk.
func (c *ConfigStore) Save(cfg interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	return c.fs.WriteFile(c.filePath, data, 0644)
}
