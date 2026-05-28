package component

import "sync"

// StateComponent 跟踪文件路径到大小的映射。
type StateComponent struct {
	mu    sync.RWMutex
	sizes map[string]int64
}

// NewStateComponent 创建一个新的 StateComponent。
func NewStateComponent() *StateComponent {
	return &StateComponent{
		sizes: make(map[string]int64),
	}
}

// GetSize 返回已记录的大小以及该路径是否存在。
func (s *StateComponent) GetSize(path string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	size, ok := s.sizes[path]
	return size, ok
}

// SetSize 记录或更新文件大小。
func (s *StateComponent) SetSize(path string, size int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sizes[path] = size
}

// Remove 从跟踪中删除指定路径。
func (s *StateComponent) Remove(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sizes, path)
}
