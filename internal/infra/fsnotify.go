package infra

import "github.com/fsnotify/fsnotify"

// FSNotifyWatcher wraps fsnotify.Watcher as infrastructure layer.
type FSNotifyWatcher struct {
	watcher *fsnotify.Watcher
}

// NewWatcher creates a new fsnotify watcher.
func NewWatcher() (*FSNotifyWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FSNotifyWatcher{watcher: w}, nil
}

// Add registers a path for monitoring.
func (f *FSNotifyWatcher) Add(path string) error {
	return f.watcher.Add(path)
}

// Remove stops monitoring a path.
func (f *FSNotifyWatcher) Remove(path string) error {
	return f.watcher.Remove(path)
}

// Events returns the event channel.
func (f *FSNotifyWatcher) Events() <-chan fsnotify.Event {
	return f.watcher.Events
}

// Errors returns the error channel.
func (f *FSNotifyWatcher) Errors() <-chan error {
	return f.watcher.Errors
}

// Close releases the watcher.
func (f *FSNotifyWatcher) Close() error {
	return f.watcher.Close()
}
