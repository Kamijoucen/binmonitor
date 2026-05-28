package logic

import (
	"fmt"
	"os"

	"binmonitor/internal/appctx"
	"binmonitor/internal/types"

	"github.com/fsnotify/fsnotify"
)

// EventRecord 承载处理文件事件的结果。
type EventRecord struct {
	Op      string
	Path    string
	OldSize int64
	NewSize int64
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
// 可能通过 appCtx 产生副作用（状态与监听器更新）。
// 返回 nil 表示无需打印记录。
func ProcessEvent(appCtx *appctx.AppCtx, event types.FileEvent) *EventRecord {
	state := appCtx.State()
	if appCtx.Ignore().ShouldIgnore(event.Path) {
		if (event.Op == types.OpRemove || event.Op == types.OpRename) && appCtx.Watcher() != nil && appCtx.Watcher().IsWatched(event.Path) {
			_ = appCtx.Watcher().RemovePath(event.Path)
		}
		return nil
	}
	shouldOutput := appCtx.EventFilter().ShouldWatch(event.Op)

	switch event.Op {
	case types.OpCreate:
		// 如果创建了新目录，确保对其进行监听并跳过记录。
		if info, err := os.Stat(event.Path); err == nil && info.IsDir() {
			if appCtx.Watcher() != nil {
				_ = appCtx.Watcher().AddPath(event.Path)
			}
			if appCtx.ReadWatcher() != nil {
				_ = appCtx.ReadWatcher().AddPath(event.Path)
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
		if appCtx.Watcher() != nil && appCtx.Watcher().IsWatched(event.Path) {
			_ = appCtx.Watcher().RemovePath(event.Path)
		}
		if appCtx.ReadWatcher() != nil {
			_ = appCtx.ReadWatcher().RemovePath(event.Path)
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
