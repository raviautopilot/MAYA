package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
)

// MockFileSystem is an in-memory filesystem for testing.
type MockFileSystem struct {
	mu    sync.RWMutex
	Files map[string][]byte
	Dirs  map[string]bool
}

func NewMockFS() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string][]byte),
		Dirs:  make(map[string]bool),
	}
}

func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.Files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return data, nil
}

func (m *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Files[path] = data
	return nil
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Dirs[path] = true
	return nil
}

func (m *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.Files[path]; ok {
		return nil, nil
	}
	return nil, os.ErrNotExist
}

type TestItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestNewStore_CreatesFile(t *testing.T) {
	fs := NewMockFS()
	store, err := NewStore[TestItem]("/tmp/test", "items.json", fs)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	if store == nil {
		t.Fatal("store is nil")
	}
	// Verify file was created
	data, err := fs.ReadFile("/tmp/test/items.json")
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if string(data) != "[]" {
		t.Errorf("expected '[]', got '%s'", string(data))
	}
}

func TestNewStore_ExistingFile(t *testing.T) {
	fs := NewMockFS()
	existing := `[{"id":"1","name":"test"}]`
	fs.Files["/tmp/test/items.json"] = []byte(existing)
	store, err := NewStore[TestItem]("/tmp/test", "items.json", fs)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	items, err := store.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
	if items[0].Name != "test" {
		t.Errorf("expected name 'test', got '%s'", items[0].Name)
	}
}

func TestStore_SaveAndLoad(t *testing.T) {
	fs := NewMockFS()
	store, _ := NewStore[TestItem]("/tmp/test", "items.json", fs)

	items := []TestItem{
		{ID: "1", Name: "Alpha"},
		{ID: "2", Name: "Beta"},
	}
	if err := store.SaveAll(items); err != nil {
		t.Fatalf("SaveAll failed: %v", err)
	}

	loaded, err := store.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	if len(loaded) != 2 {
		t.Errorf("expected 2 items, got %d", len(loaded))
	}
}

func TestStore_WithLock(t *testing.T) {
	fs := NewMockFS()
	store, _ := NewStore[TestItem]("/tmp/test", "items.json", fs)

	// Add item via WithLock
	err := store.WithLock(func(items []TestItem) ([]TestItem, error) {
		return append(items, TestItem{ID: "1", Name: "New"}), nil
	})
	if err != nil {
		t.Fatalf("WithLock failed: %v", err)
	}

	loaded, _ := store.LoadAll()
	if len(loaded) != 1 {
		t.Errorf("expected 1 item, got %d", len(loaded))
	}
}

func TestStore_WithLock_Error(t *testing.T) {
	fs := NewMockFS()
	store, _ := NewStore[TestItem]("/tmp/test", "items.json", fs)

	err := store.WithLock(func(items []TestItem) ([]TestItem, error) {
		return nil, fmt.Errorf("test error")
	})
	if err == nil {
		t.Fatal("expected error from WithLock")
	}
}

func TestStore_ConcurrentAccess(t *testing.T) {
	fs := NewMockFS()
	store, _ := NewStore[TestItem]("/tmp/test", "items.json", fs)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = store.WithLock(func(items []TestItem) ([]TestItem, error) {
				return append(items, TestItem{ID: fmt.Sprintf("%d", n), Name: "test"}), nil
			})
		}(i)
	}
	wg.Wait()

	loaded, _ := store.LoadAll()
	if len(loaded) != 50 {
		t.Errorf("expected 50 items after concurrent writes, got %d", len(loaded))
	}
}

func TestConfigStore_LoadSave(t *testing.T) {
	fs := NewMockFS()
	cfg := map[string]interface{}{"port": 8080, "name": "test"}
	data, _ := json.Marshal(cfg)
	fs.Files["/tmp/config.json"] = data

	cs := NewConfigStore("/tmp/config.json", fs)
	var loaded map[string]interface{}
	if err := cs.Load(&loaded); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded["name"] != "test" {
		t.Errorf("expected name 'test', got %v", loaded["name"])
	}

	// Save
	loaded["name"] = "updated"
	if err := cs.Save(loaded); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	var reloaded map[string]interface{}
	_ = cs.Load(&reloaded)
	if reloaded["name"] != "updated" {
		t.Errorf("expected name 'updated', got %v", reloaded["name"])
	}
}
