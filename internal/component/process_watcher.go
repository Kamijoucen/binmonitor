package component

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"binmonitor/internal/types"
)

type processFDSnapshot map[int]string

// ProcessWatcherComponent 监听单个进程的文件描述符变化。
type ProcessWatcherComponent struct {
	process   types.ProcessConfig
	interval  time.Duration
	events    chan types.FileEvent
	errors    chan error
	done      chan struct{}
	closeOnce sync.Once
}

// MultiProcessWatcherComponent 聚合多个进程监听器的事件。
type MultiProcessWatcherComponent struct {
	watchers  []*ProcessWatcherComponent
	events    chan types.FileEvent
	errors    chan error
	done      chan struct{}
	closeOnce sync.Once
}

func newProcessWatcherComponent(process types.ProcessConfig, interval time.Duration) *ProcessWatcherComponent {
	return &ProcessWatcherComponent{
		process:  process,
		interval: interval,
		events:   make(chan types.FileEvent, 128),
		errors:   make(chan error, 64),
		done:     make(chan struct{}),
	}
}

// NewMultiProcessWatcherComponent 创建多进程监听聚合组件。
func NewMultiProcessWatcherComponent(processes []types.ProcessConfig) (*MultiProcessWatcherComponent, error) {
	if len(processes) == 0 {
		return nil, fmt.Errorf("process watcher requires at least one process")
	}
	watchers := make([]*ProcessWatcherComponent, 0, len(processes))
	for _, process := range processes {
		watcher, err := NewProcessWatcherComponent(process)
		if err != nil {
			return nil, err
		}
		watchers = append(watchers, watcher)
	}
	return &MultiProcessWatcherComponent{
		watchers: watchers,
		events:   make(chan types.FileEvent, 128),
		errors:   make(chan error, 64),
		done:     make(chan struct{}),
	}, nil
}

// Events 返回聚合后的进程文件事件通道。
func (w *MultiProcessWatcherComponent) Events() <-chan types.FileEvent {
	return w.events
}

// Errors 返回聚合后的错误通道。
func (w *MultiProcessWatcherComponent) Errors() <-chan error {
	return w.errors
}

// Close 停止所有进程监听。
func (w *MultiProcessWatcherComponent) Close() error {
	w.closeOnce.Do(func() {
		close(w.done)
		for _, watcher := range w.watchers {
			_ = watcher.Close()
		}
	})
	return nil
}

// Run 启动所有子监听器并转发事件。
func (w *MultiProcessWatcherComponent) Run() {
	var wg sync.WaitGroup
	for _, watcher := range w.watchers {
		watcher := watcher
		wg.Add(3)
		go func() {
			defer wg.Done()
			watcher.Run()
		}()
		go func() {
			defer wg.Done()
			for event := range watcher.Events() {
				select {
				case w.events <- event:
				case <-w.done:
					return
				}
			}
		}()
		go func() {
			defer wg.Done()
			for err := range watcher.Errors() {
				select {
				case w.errors <- err:
				case <-w.done:
					return
				}
			}
		}()
	}
	wg.Wait()
	close(w.events)
	close(w.errors)
}

// Events 返回单进程文件事件通道。
func (w *ProcessWatcherComponent) Events() <-chan types.FileEvent {
	return w.events
}

// Errors 返回单进程错误通道。
func (w *ProcessWatcherComponent) Errors() <-chan error {
	return w.errors
}

// Close 停止单进程监听。
func (w *ProcessWatcherComponent) Close() error {
	w.closeOnce.Do(func() {
		close(w.done)
	})
	return nil
}

func (w *ProcessWatcherComponent) sendEvent(event types.FileEvent) bool {
	select {
	case w.events <- event:
		return true
	case <-w.done:
		return false
	}
}

func (w *ProcessWatcherComponent) sendError(err error) bool {
	select {
	case w.errors <- err:
		return true
	case <-w.done:
		return false
	}
}

func diffProcessFDSnapshot(pid int, processName string, previous processFDSnapshot, next processFDSnapshot) []types.FileEvent {
	fds := make(map[int]struct{}, len(previous)+len(next))
	for fd := range previous {
		fds[fd] = struct{}{}
	}
	for fd := range next {
		fds[fd] = struct{}{}
	}

	orderedFDs := make([]int, 0, len(fds))
	for fd := range fds {
		orderedFDs = append(orderedFDs, fd)
	}
	sort.Ints(orderedFDs)

	events := make([]types.FileEvent, 0)
	for _, fd := range orderedFDs {
		oldPath, hadOld := previous[fd]
		newPath, hasNew := next[fd]
		if hadOld && (!hasNew || oldPath != newPath) {
			events = append(events, types.FileEvent{Path: oldPath, Op: types.OpClose, PID: pid, FD: fd, ProcessName: processName})
		}
		if hasNew && (!hadOld || oldPath != newPath) {
			events = append(events, types.FileEvent{Path: newPath, Op: types.OpOpen, PID: pid, FD: fd, ProcessName: processName})
		}
	}
	return events
}

func isProcessFileTarget(target string) bool {
	if target == "" || !strings.HasPrefix(target, "/") {
		return false
	}
	for _, prefix := range []string{"socket:", "pipe:", "anon_inode:", "memfd:"} {
		if strings.HasPrefix(target, prefix) {
			return false
		}
	}
	return true
}
