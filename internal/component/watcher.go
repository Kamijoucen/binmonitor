package component

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// WatcherComponent 管理递归目录监听并暴露原始 fsnotify 事件。
type WatcherComponent struct {
	watcher *fsnotify.Watcher
	paths   map[string]struct{}
	mu      sync.RWMutex
	events  chan fsnotify.Event
	errors  chan error
	done    chan struct{}
}

// NewWatcherComponent 创建一个 WatcherComponent。
func NewWatcherComponent() (*WatcherComponent, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &WatcherComponent{
		watcher: w,
		paths:   make(map[string]struct{}),
		events:  make(chan fsnotify.Event, 128),
		errors:  make(chan error, 64),
		done:    make(chan struct{}),
	}, nil
}

// AddRecursive 遍历 root 并将所有目录加入监听。
func (w *WatcherComponent) AddRecursive(root string) error {
	return w.AddRecursiveWithFilter(root, nil)
}

// AddRecursiveWithFilter 遍历 root，并将未被过滤的目录加入监听。
func (w *WatcherComponent) AddRecursiveWithFilter(root string, shouldSkip func(path string, info os.FileInfo) bool) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if shouldSkip != nil && shouldSkip(path, info) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			if addErr := w.watcher.Add(path); addErr == nil {
				w.mu.Lock()
				w.paths[path] = struct{}{}
				w.mu.Unlock()
			}
		}
		return nil
	})
}

// AddPath 添加单个目录路径。
func (w *WatcherComponent) AddPath(path string) error {
	if err := w.watcher.Add(path); err != nil {
		return err
	}
	w.mu.Lock()
	w.paths[path] = struct{}{}
	w.mu.Unlock()
	return nil
}

// RemovePath 移除一个目录路径。
func (w *WatcherComponent) RemovePath(path string) error {
	if err := w.watcher.Remove(path); err != nil {
		return err
	}
	w.mu.Lock()
	delete(w.paths, path)
	w.mu.Unlock()
	return nil
}

// Events 返回原始 fsnotify 事件通道。
func (w *WatcherComponent) Events() <-chan fsnotify.Event {
	return w.events
}

// Errors 返回错误通道。
func (w *WatcherComponent) Errors() <-chan error {
	return w.errors
}

// IsWatched 报告某个目录是否正在被监听。
func (w *WatcherComponent) IsWatched(path string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	_, ok := w.paths[path]
	return ok
}

// Close 停止监听并关闭通道。
func (w *WatcherComponent) Close() error {
	close(w.done)
	return w.watcher.Close()
}

// Run 开始转发 fsnotify 事件，会阻塞直到监听器被关闭。
func (w *WatcherComponent) Run() {
	for {
		select {
		case <-w.done:
			return
		case ev, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			select {
			case w.events <- ev:
			case <-w.done:
				return
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			select {
			case w.errors <- err:
			case <-w.done:
				return
			}
		}
	}
}
