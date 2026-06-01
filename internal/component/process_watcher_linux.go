//go:build linux

package component

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"binmonitor/internal/types"
)

// NewProcessWatcherComponent 创建单个 Linux 进程文件描述符监听组件。
func NewProcessWatcherComponent(process types.ProcessConfig) (*ProcessWatcherComponent, error) {
	if process.PID <= 0 {
		return nil, fmt.Errorf("process pid must be greater than 0")
	}
	if process.PollIntervalMs <= 0 {
		process.PollIntervalMs = types.DefaultProcessPollIntervalMs
	}
	return newProcessWatcherComponent(process, time.Duration(process.PollIntervalMs)*time.Millisecond), nil
}

// Run 开始轮询目标进程的文件描述符快照。
func (w *ProcessWatcherComponent) Run() {
	defer close(w.events)
	defer close(w.errors)

	previous, err := snapshotProcessFDs(w.process.PID)
	if err != nil {
		w.sendError(fmt.Errorf("snapshot process %d fds: %w", w.process.PID, err))
		return
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-ticker.C:
			next, err := snapshotProcessFDs(w.process.PID)
			if err != nil {
				w.sendError(fmt.Errorf("snapshot process %d fds: %w", w.process.PID, err))
				return
			}
			for _, event := range diffProcessFDSnapshot(w.process.PID, w.process.Name, previous, next) {
				if !w.sendEvent(event) {
					return
				}
			}
			previous = next
		}
	}
}

func snapshotProcessFDs(pid int) (processFDSnapshot, error) {
	fdDir := filepath.Join("/proc", strconv.Itoa(pid), "fd")
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return nil, err
	}

	snapshot := make(processFDSnapshot, len(entries))
	for _, entry := range entries {
		fd, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		fdPath := filepath.Join(fdDir, entry.Name())
		target, err := os.Readlink(fdPath)
		if err != nil || !isProcessFileTarget(target) {
			continue
		}
		info, err := os.Stat(fdPath)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		snapshot[fd] = target
	}
	return snapshot, nil
}
