package logic

import (
	"fmt"
	"os"

	"binmonitor/internal/types"

	"github.com/fsnotify/fsnotify"
)

// ─── 能力接口（由 Logic 自身定义，由 Component 实现） ───

// FileStateTracker 定义跟踪文件大小的能力。
type FileStateTracker interface {
	GetSize(path string) (int64, bool)
	SetSize(path string, size int64)
	Remove(path string)
}

// DirectoryWatcher 定义管理目录监听的能力。
type DirectoryWatcher interface {
	IsWatched(path string) bool
	RemovePath(path string) error
	AddPath(path string) error
}

// EventTypeFilter 定义过滤文件操作类型的能力。
type EventTypeFilter interface {
	ShouldWatch(op types.FileOp) bool
}

// ReadDirectoryWatcher 定义管理读取监听目录的能力。
type ReadDirectoryWatcher interface {
	AddPath(path string) error
	RemovePath(path string) error
}

// EventRecord 承载处理文件事件的结果。
type EventRecord struct {
	Op             string
	Path           string
	OldSize        int64
	NewSize        int64
	PID            int
	FD             int
	ProcessName    string
	HasProcessMeta bool
}

// TranslateEvent 将原始 fsnotify 事件转换为领域 FileEvent。
func TranslateEvent(ev fsnotify.Event) *types.FileEvent {
	var op types.FileOp
	switch {
	case ev.Op&fsnotify.Create != 0:
		op = types.OpCreate
	case ev.Op&fsnotify.Write != 0:
		op = types.OpWrite
	case ev.Op&fsnotify.Remove != 0:
		op = types.OpRemove
	case ev.Op&fsnotify.Rename != 0:
		op = types.OpRename
	case ev.Op&fsnotify.Chmod != 0:
		op = types.OpChmod
	default:
		return nil
	}
	return &types.FileEvent{Path: ev.Name, Op: op}
}

// ProcessEvent 处理文件事件并返回用于输出的 EventRecord。
// 副作用（状态更新与监听器管理）通过注入的接口实例完成。
// 返回 nil 表示无需打印记录。
func ProcessEvent(state FileStateTracker, ignore PathIgnorer, watcher DirectoryWatcher, eventFilter EventTypeFilter, readWatcher ReadDirectoryWatcher, event types.FileEvent) *EventRecord {
	if event.Op == types.OpOpen || event.Op == types.OpClose {
		return processProcessEvent(eventFilter, event)
	}

	if ignore.ShouldIgnore(event.Path) {
		if (event.Op == types.OpRemove || event.Op == types.OpRename) && watcher != nil && watcher.IsWatched(event.Path) {
			_ = watcher.RemovePath(event.Path)
		}
		return nil
	}
	shouldOutput := eventFilter.ShouldWatch(event.Op)

	switch event.Op {
	case types.OpCreate:
		// 如果创建了新目录，确保对其进行监听并跳过记录。
		if info, err := os.Stat(event.Path); err == nil && info.IsDir() {
			if watcher != nil {
				_ = watcher.AddPath(event.Path)
			}
			if readWatcher != nil {
				_ = readWatcher.AddPath(event.Path)
			}
			return nil
		}
		newSize := getFileSize(event.Path)
		state.SetSize(event.Path, newSize)
		if !shouldOutput {
			return nil
		}
		return &EventRecord{Op: "CREATE", Path: event.Path, OldSize: 0, NewSize: newSize}

	case types.OpWrite:
		oldSize, _ := state.GetSize(event.Path)
		newSize := getFileSize(event.Path)
		state.SetSize(event.Path, newSize)
		if !shouldOutput {
			return nil
		}
		return &EventRecord{Op: "WRITE", Path: event.Path, OldSize: oldSize, NewSize: newSize}

	case types.OpRemove, types.OpRename:
		oldSize, _ := state.GetSize(event.Path)
		state.Remove(event.Path)
		opName := "REMOVE"
		if event.Op == types.OpRename {
			opName = "RENAME"
		}

		// 如果目录被移除，停止监听它。
		if watcher != nil && watcher.IsWatched(event.Path) {
			_ = watcher.RemovePath(event.Path)
		}
		if readWatcher != nil {
			_ = readWatcher.RemovePath(event.Path)
		}
		if !shouldOutput {
			return nil
		}
		return &EventRecord{Op: opName, Path: event.Path, OldSize: oldSize, NewSize: 0}

	case types.OpRead:
		if !shouldOutput {
			return nil
		}
		size := getFileSize(event.Path)
		return &EventRecord{Op: "READ", Path: event.Path, OldSize: size, NewSize: size}

	case types.OpChmod:
		// 无大小变化，忽略。
		return nil
	}

	return nil
}

func processProcessEvent(eventFilter EventTypeFilter, event types.FileEvent) *EventRecord {
	if !eventFilter.ShouldWatch(event.Op) {
		return nil
	}
	opName := "OPEN"
	if event.Op == types.OpClose {
		opName = "CLOSE"
	}
	size := getFileSize(event.Path)
	return &EventRecord{
		Op:             opName,
		Path:           event.Path,
		OldSize:        size,
		NewSize:        size,
		PID:            event.PID,
		FD:             event.FD,
		ProcessName:    event.ProcessName,
		HasProcessMeta: true,
	}
}

func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	if info.IsDir() {
		return 0
	}
	return info.Size()
}

// ProcessNetEvent 处理网络连接事件并返回用于输出的 EventRecord。
func ProcessNetEvent(eventFilter EventTypeFilter, event types.NetEvent) *EventRecord {
	_ = eventFilter // 保持接口一致，暂不通过 EventFilter 过滤网络事件

	return &EventRecord{
		Op:             event.Op.String(),
		Path:           event.ConnectionString(),
		PID:            event.PID,
		FD:             event.FD,
		ProcessName:    event.ProcessName,
		HasProcessMeta: true,
	}
}

// HumanReadableSize 将字节数转换为人类可读的字符串。
func HumanReadableSize(bytes int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if bytes < 0 {
		return "-" + HumanReadableSize(-bytes)
	}

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2fGB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
