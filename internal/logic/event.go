package logic

import (
	"fmt"
	"os"
	"time"

	"binmonitor/internal/appctx"
	"binmonitor/internal/component"
)

// ProcessEvent handles a file event, computes size changes, and prints a record.
func ProcessEvent(appCtx *appctx.AppCtx, event component.FileEvent) {
	state := appCtx.State()

	switch event.Op {
	case component.OpCreate:
		// If a new directory is created, ensure it is watched and skip recording.
		if info, err := os.Stat(event.Path); err == nil && info.IsDir() {
			_ = appCtx.Watcher().AddPath(event.Path)
			return
		}
		newSize := getFileSize(event.Path)
		state.SetSize(event.Path, newSize)
		printRecord("CREATE", event.Path, 0, newSize)

	case component.OpWrite:
		oldSize, _ := state.GetSize(event.Path)
		newSize := getFileSize(event.Path)
		state.SetSize(event.Path, newSize)
		printRecord("WRITE", event.Path, oldSize, newSize)

	case component.OpRemove, component.OpRename:
		oldSize, _ := state.GetSize(event.Path)
		state.Remove(event.Path)
		opName := "REMOVE"
		if event.Op == component.OpRename {
			opName = "RENAME"
		}
		printRecord(opName, event.Path, oldSize, 0)

		// If a directory is removed, stop watching it.
		if appCtx.Watcher().IsWatched(event.Path) {
			_ = appCtx.Watcher().RemovePath(event.Path)
		}

	case component.OpChmod:
		// No size change; ignore.
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

func printRecord(op, path string, oldSize, newSize int64) {
	diff := newSize - oldSize
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%s %s %s %s → %s (%+s)\n",
		timestamp, op, path,
		HumanReadableSize(oldSize),
		HumanReadableSize(newSize),
		HumanReadableSize(diff),
	)
}

// HumanReadableSize converts bytes to a human-readable string.
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
