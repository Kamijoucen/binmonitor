package component

import (
	"os"
	"path/filepath"
	"sync"
)

// StateComponent tracks file path to size mapping.
type StateComponent struct {
	mu     sync.RWMutex
	sizes  map[string]int64
}

// NewStateComponent creates a new StateComponent.
func NewStateComponent() *StateComponent {
	return &StateComponent{
		sizes: make(map[string]int64),
	}
}

// InitFromPath recursively scans root and records all file sizes.
func (s *StateComponent) InitFromPath(root string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Permission errors or other walk errors: skip but continue
			return nil
		}
		if !info.IsDir() {
			s.sizes[path] = info.Size()
		}
		return nil
	})
}

// GetSize returns the recorded size and whether the path exists.
func (s *StateComponent) GetSize(path string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	size, ok := s.sizes[path]
	return size, ok
}

// SetSize records or updates a file's size.
func (s *StateComponent) SetSize(path string, size int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sizes[path] = size
}

// Remove deletes a path from tracking.
func (s *StateComponent) Remove(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sizes, path)
}
