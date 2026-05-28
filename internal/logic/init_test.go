package logic

import (
	"os"
	"path/filepath"
	"testing"

	"binmonitor/internal/appctx"
	"binmonitor/internal/component"
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
	eventFilter, err := component.NewEventFilterComponent(DefaultConfig().Events)
	if err != nil {
		t.Fatalf("NewEventFilterComponent() error = %v", err)
	}
	applicationContext := appctx.NewAppCtx(nil, state, ignore, eventFilter, nil, nil)

	if err := InitStateFromPath(applicationContext, root); err != nil {
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
	eventFilter, err := component.NewEventFilterComponent(DefaultConfig().Events)
	if err != nil {
		t.Fatalf("NewEventFilterComponent() error = %v", err)
	}
	applicationContext := appctx.NewAppCtx(nil, state, ignore, eventFilter, nil, nil)

	if record := ProcessEvent(applicationContext, types.FileEvent{Path: ignoredPath, Op: types.OpWrite}); record != nil {
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
	eventFilter, err := component.NewEventFilterComponent([]string{"remove"})
	if err != nil {
		t.Fatalf("NewEventFilterComponent() error = %v", err)
	}
	applicationContext := appctx.NewAppCtx(nil, state, ignore, eventFilter, nil, nil)

	if record := ProcessEvent(applicationContext, types.FileEvent{Path: path, Op: types.OpWrite}); record != nil {
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
	eventFilter, err := component.NewEventFilterComponent([]string{"read"})
	if err != nil {
		t.Fatalf("NewEventFilterComponent() error = %v", err)
	}
	applicationContext := appctx.NewAppCtx(nil, state, ignore, eventFilter, nil, nil)

	record := ProcessEvent(applicationContext, types.FileEvent{Path: path, Op: types.OpRead})
	if record == nil {
		t.Fatal("ProcessEvent() = nil, want READ record")
	}
	if record.Op != "READ" || record.OldSize != 5 || record.NewSize != 5 {
		t.Fatalf("record = %#v, want READ with size 5", record)
	}
}
