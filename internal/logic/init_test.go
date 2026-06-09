package logic_test

import (
	"os"
	"path/filepath"
	"testing"

	"binmonitor/internal/component"
	"binmonitor/internal/logic"
	"binmonitor/internal/types"
)

func TestInitStateFromPathSkipsIgnoredFilesAndDirs(t *testing.T) {
	root := t.TempDir()
	keepPath := filepath.Join(root, "keep.txt")
	ignoredFilePath := filepath.Join(root, "tmp", "cache.db")
	ignoredDirFilePath := filepath.Join(root, "logs", "app.log")

	if err := os.MkdirAll(filepath.Dir(ignoredFilePath), 0755); err != nil {
		t.Fatalf("mkdir ignored file dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(ignoredDirFilePath), 0755); err != nil {
		t.Fatalf("mkdir ignored dir: %v", err)
	}
	if err := os.WriteFile(keepPath, []byte("keep"), 0644); err != nil {
		t.Fatalf("write keep file: %v", err)
	}
	if err := os.WriteFile(ignoredFilePath, []byte("ignored"), 0644); err != nil {
		t.Fatalf("write ignored file: %v", err)
	}
	if err := os.WriteFile(ignoredDirFilePath, []byte("ignored"), 0644); err != nil {
		t.Fatalf("write ignored dir file: %v", err)
	}

	state := component.NewStateComponent()
	ignore := component.NewIgnoreComponent(root, []string{"tmp/cache.db", "logs"})

	if err := logic.InitStateFromPath(state, ignore, root); err != nil {
		t.Fatalf("InitStateFromPath() error = %v", err)
	}
	if _, ok := state.GetSize(keepPath); !ok {
		t.Fatal("keep file was not tracked")
	}
	if _, ok := state.GetSize(ignoredFilePath); ok {
		t.Fatal("ignored file was tracked")
	}
	if _, ok := state.GetSize(ignoredDirFilePath); ok {
		t.Fatal("ignored dir file was tracked")
	}
}

func TestProcessEventSkipsIgnoredPath(t *testing.T) {
	root := t.TempDir()
	ignoredPath := filepath.Join(root, "logs", "app.log")
	if err := os.MkdirAll(filepath.Dir(ignoredPath), 0755); err != nil {
		t.Fatalf("mkdir ignored dir: %v", err)
	}
	if err := os.WriteFile(ignoredPath, []byte("ignored"), 0644); err != nil {
		t.Fatalf("write ignored file: %v", err)
	}

	state := component.NewStateComponent()
	ignore := component.NewIgnoreComponent(root, []string{"logs"})
	ops, err := logic.ResolveAllEventOps(logic.DefaultConfig().Events)
	if err != nil {
		t.Fatalf("ResolveAllEventOps() error = %v", err)
	}
	eventFilter := component.NewEventFilterComponent(ops)

	if record := logic.ProcessEvent(state, ignore, nil, eventFilter, nil, types.FileEvent{Path: ignoredPath, Op: types.OpWrite}); record != nil {
		t.Fatalf("ProcessEvent() = %#v, want nil", record)
	}
	if _, ok := state.GetSize(ignoredPath); ok {
		t.Fatal("ignored event updated state")
	}
}

func TestProcessEventSkipsDisabledEventOutputButUpdatesState(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	state := component.NewStateComponent()
	ignore := component.NewIgnoreComponent(root, nil)
	ops, err := logic.ResolveAllEventOps([]string{"remove"})
	if err != nil {
		t.Fatalf("ResolveAllEventOps() error = %v", err)
	}
	eventFilter := component.NewEventFilterComponent(ops)

	if record := logic.ProcessEvent(state, ignore, nil, eventFilter, nil, types.FileEvent{Path: path, Op: types.OpWrite}); record != nil {
		t.Fatalf("ProcessEvent() = %#v, want nil", record)
	}
	if size, ok := state.GetSize(path); !ok || size != 5 {
		t.Fatalf("state size = %d, ok = %v, want 5 and true", size, ok)
	}
}

func TestProcessEventReturnsReadRecordWhenEnabled(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	state := component.NewStateComponent()
	ignore := component.NewIgnoreComponent(root, nil)
	ops, err := logic.ResolveAllEventOps([]string{"read"})
	if err != nil {
		t.Fatalf("ResolveAllEventOps() error = %v", err)
	}
	eventFilter := component.NewEventFilterComponent(ops)

	record := logic.ProcessEvent(state, ignore, nil, eventFilter, nil, types.FileEvent{Path: path, Op: types.OpRead})
	if record == nil {
		t.Fatal("ProcessEvent() = nil, want READ record")
	}
	if record.Op != "READ" || record.OldSize != 5 || record.NewSize != 5 {
		t.Fatalf("record = %#v, want READ with size 5", record)
	}
}

func TestProcessEventReturnsOpenRecordWithProcessMetadata(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	state := component.NewStateComponent()
	ignore := component.NewIgnoreComponent(root, []string{"file.txt"})
	ops, err := logic.ResolveAllEventOps([]string{"open"})
	if err != nil {
		t.Fatalf("ResolveAllEventOps() error = %v", err)
	}
	eventFilter := component.NewEventFilterComponent(ops)

	record := logic.ProcessEvent(state, ignore, nil, eventFilter, nil, types.FileEvent{Path: path, Op: types.OpOpen, PID: 123, FD: 7, ProcessName: "worker"})
	if record == nil {
		t.Fatal("ProcessEvent() = nil, want OPEN record")
	}
	if record.Op != "OPEN" || record.PID != 123 || record.FD != 7 || record.ProcessName != "worker" || !record.HasProcessMeta {
		t.Fatalf("record = %#v, want process metadata", record)
	}
	if record.NewSize != 5 || record.OldSize != 5 {
		t.Fatalf("record sizes = %d/%d, want 5/5", record.OldSize, record.NewSize)
	}

	line := logic.FormatEventRecord(record, "2026-06-01 12:00:00")
	wantLine := "2026-06-01 12:00:00 OPEN pid=123 fd=7 name=worker " + path + " 5B"
	if line != wantLine {
		t.Fatalf("FormatEventRecord() = %q, want %q", line, wantLine)
	}
	if got := logic.DedupRecordPath(record); got != "pid=123 name=worker "+path {
		t.Fatalf("DedupRecordPath() = %q", got)
	}
}
