package component

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// WatcherComponent manages recursive directory watching and translates
// fsnotify events into domain FileEvents.
type WatcherComponent struct {
	watcher *fsnotify.Watcher
	paths   map[string]struct{}
	mu      sync.RWMutex
	events  chan FileEvent
	errors  chan error
	done    chan struct{}
}

// NewWatcherComponent creates a WatcherComponent.
func NewWatcherComponent() (*WatcherComponent, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &WatcherComponent{
		watcher: w,
		paths:   make(map[string]struct{}),
		events:  make(chan FileEvent, 128),
		errors:  make(chan error, 64),
		done:    make(chan struct{}),
	}, nil
}

// AddRecursive walks root and adds all directories to the watcher.
func (w *WatcherComponent) AddRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
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

// AddPath adds a single directory path.
func (w *WatcherComponent) AddPath(path string) error {
	if err := w.watcher.Add(path); err != nil {
		return err
	}
	w.mu.Lock()
	w.paths[path] = struct{}{}
	w.mu.Unlock()
	return nil
}

// RemovePath removes a directory path.
func (w *WatcherComponent) RemovePath(path string) error {
	if err := w.watcher.Remove(path); err != nil {
		return err
	}
	w.mu.Lock()
	delete(w.paths, path)
	w.mu.Unlock()
	return nil
}

// Events returns the domain event channel.
func (w *WatcherComponent) Events() <-chan FileEvent {
	return w.events
}

// Errors returns the error channel.
func (w *WatcherComponent) Errors() <-chan error {
	return w.errors
}

// IsWatched reports whether a directory is being watched.
func (w *WatcherComponent) IsWatched(path string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	_, ok := w.paths[path]
	return ok
}

// Close stops the watcher and closes channels.
func (w *WatcherComponent) Close() error {
	close(w.done)
	return w.watcher.Close()
}

// Run starts translating infrastructure events into FileEvents.
// It blocks until the watcher is closed.
func (w *WatcherComponent) Run() {
	for {
		select {
		case <-w.done:
			return
		case ev, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			fe := w.translate(ev)
			if fe != nil {
				select {
				case w.events <- *fe:
				case <-w.done:
					return
				}
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

func (w *WatcherComponent) translate(ev fsnotify.Event) *FileEvent {
	var op FileOp
	switch {
	case ev.Op&fsnotify.Create != 0:
		op = OpCreate
	case ev.Op&fsnotify.Write != 0:
		op = OpWrite
	case ev.Op&fsnotify.Remove != 0:
		op = OpRemove
	case ev.Op&fsnotify.Rename != 0:
		op = OpRename
	case ev.Op&fsnotify.Chmod != 0:
		op = OpChmod
	default:
		return nil
	}
	return &FileEvent{Path: ev.Name, Op: op}
}
