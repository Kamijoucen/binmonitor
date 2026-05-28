//go:build linux

package component

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"binmonitor/internal/types"

	"golang.org/x/sys/unix"
)

// ReadWatcherComponent 基于 inotify 监听文件读取事件。
type ReadWatcherComponent struct {
	fd        int
	paths     map[int]string
	watches   map[string]int
	mu        sync.RWMutex
	events    chan types.FileEvent
	errors    chan error
	done      chan struct{}
	closeOnce sync.Once
}

// NewReadWatcherComponent 创建 ReadWatcherComponent。
func NewReadWatcherComponent() (*ReadWatcherComponent, error) {
	fd, err := unix.InotifyInit1(unix.IN_CLOEXEC | unix.IN_NONBLOCK)
	if err != nil {
		return nil, err
	}
	return &ReadWatcherComponent{
		fd:      fd,
		paths:   make(map[int]string),
		watches: make(map[string]int),
		events:  make(chan types.FileEvent, 128),
		errors:  make(chan error, 64),
		done:    make(chan struct{}),
	}, nil
}

// AddRecursiveWithFilter 遍历 root，并将未被过滤的目录加入读取监听。
func (w *ReadWatcherComponent) AddRecursiveWithFilter(root string, shouldSkip func(path string, info os.FileInfo) bool) error {
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
			return w.AddPath(path)
		}
		return nil
	})
}

// AddPath 添加单个目录路径到读取监听。
func (w *ReadWatcherComponent) AddPath(path string) error {
	cleanPath := filepath.Clean(path)
	w.mu.RLock()
	_, ok := w.watches[cleanPath]
	w.mu.RUnlock()
	if ok {
		return nil
	}

	mask := uint32(unix.IN_ACCESS | unix.IN_CREATE | unix.IN_MOVED_TO | unix.IN_DELETE_SELF | unix.IN_MOVE_SELF)
	wd, err := unix.InotifyAddWatch(w.fd, cleanPath, mask)
	if err != nil {
		return err
	}
	w.mu.Lock()
	w.paths[wd] = cleanPath
	w.watches[cleanPath] = wd
	w.mu.Unlock()
	return nil
}

// RemovePath 移除一个目录路径的读取监听。
func (w *ReadWatcherComponent) RemovePath(path string) error {
	cleanPath := filepath.Clean(path)
	w.mu.Lock()
	wd, ok := w.watches[cleanPath]
	if ok {
		delete(w.watches, cleanPath)
		delete(w.paths, wd)
	}
	w.mu.Unlock()
	if !ok {
		return nil
	}
	_, err := unix.InotifyRmWatch(w.fd, uint32(wd))
	return err
}

// Events 返回读取事件通道。
func (w *ReadWatcherComponent) Events() <-chan types.FileEvent {
	return w.events
}

// Errors 返回错误通道。
func (w *ReadWatcherComponent) Errors() <-chan error {
	return w.errors
}

// Close 停止读取监听。
func (w *ReadWatcherComponent) Close() error {
	var closeErr error
	w.closeOnce.Do(func() {
		close(w.done)
		closeErr = unix.Close(w.fd)
	})
	return closeErr
}

// Run 开始转发读取事件，会阻塞直到监听器被关闭。
func (w *ReadWatcherComponent) Run() {
	buffer := make([]byte, 4096)
	for {
		select {
		case <-w.done:
			return
		default:
		}

		n, err := unix.Read(w.fd, buffer)
		if err != nil {
			if errors.Is(err, unix.EAGAIN) {
				if !w.waitReadable() {
					return
				}
				continue
			}
			if errors.Is(err, unix.EINTR) {
				continue
			}
			if errors.Is(err, unix.EBADF) {
				return
			}
			w.sendError(err)
			return
		}
		if n > 0 {
			w.handleEvents(buffer[:n])
		}
	}
}

func (w *ReadWatcherComponent) waitReadable() bool {
	for {
		select {
		case <-w.done:
			return false
		default:
		}

		pollFd := []unix.PollFd{{Fd: int32(w.fd), Events: unix.POLLIN}}
		n, err := unix.Poll(pollFd, 100)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if errors.Is(err, unix.EBADF) {
			return false
		}
		if err != nil {
			w.sendError(err)
			return false
		}
		if n > 0 {
			return true
		}
	}
}

func (w *ReadWatcherComponent) handleEvents(data []byte) {
	for offset := 0; offset+unix.SizeofInotifyEvent <= len(data); {
		raw := (*unix.InotifyEvent)(unsafe.Pointer(&data[offset]))
		nextOffset := offset + unix.SizeofInotifyEvent + int(raw.Len)
		if nextOffset > len(data) {
			return
		}

		name := strings.TrimRight(string(data[offset+unix.SizeofInotifyEvent:nextOffset]), "\x00")
		path := w.eventPath(int(raw.Wd), name)
		mask := uint32(raw.Mask)

		if mask&(unix.IN_IGNORED|unix.IN_DELETE_SELF|unix.IN_MOVE_SELF) != 0 {
			w.removeWatchDescriptor(int(raw.Wd))
		}
		if mask&(unix.IN_CREATE|unix.IN_MOVED_TO) != 0 && mask&unix.IN_ISDIR != 0 {
			_ = w.AddPath(path)
		}
		if mask&unix.IN_ACCESS != 0 && mask&unix.IN_ISDIR == 0 && path != "" {
			w.sendEvent(types.FileEvent{Path: path, Op: types.OpRead})
		}

		offset = nextOffset
	}
}

func (w *ReadWatcherComponent) eventPath(wd int, name string) string {
	w.mu.RLock()
	dir := w.paths[wd]
	w.mu.RUnlock()
	if dir == "" {
		return name
	}
	if name == "" {
		return dir
	}
	return filepath.Join(dir, name)
}

func (w *ReadWatcherComponent) removeWatchDescriptor(wd int) {
	w.mu.Lock()
	path := w.paths[wd]
	delete(w.paths, wd)
	if path != "" {
		delete(w.watches, path)
	}
	w.mu.Unlock()
}

func (w *ReadWatcherComponent) sendEvent(event types.FileEvent) {
	select {
	case w.events <- event:
	case <-w.done:
	}
}

func (w *ReadWatcherComponent) sendError(err error) {
	select {
	case w.errors <- err:
	case <-w.done:
	}
}
