package component

import (
	"path/filepath"
	"testing"
)

func TestIgnoreComponentMatchesFile(t *testing.T) {
	root := filepath.Join(t.TempDir(), "watch")
	ignore := NewIgnoreComponent(root, []string{"tmp/cache.db"})

	if !ignore.ShouldIgnore(filepath.Join(root, "tmp", "cache.db")) {
		t.Fatal("ShouldIgnore() = false, want true")
	}
	if ignore.ShouldIgnore(filepath.Join(root, "tmp", "cache.db.bak")) {
		t.Fatal("ShouldIgnore() = true, want false")
	}
}

func TestIgnoreComponentMatchesDirectoryChildren(t *testing.T) {
	root := filepath.Join(t.TempDir(), "watch")
	ignore := NewIgnoreComponent(root, []string{"logs"})

	if !ignore.ShouldIgnore(filepath.Join(root, "logs", "app.log")) {
		t.Fatal("ShouldIgnore() child = false, want true")
	}
	if !ignore.ShouldIgnore(filepath.Join(root, "logs")) {
		t.Fatal("ShouldIgnore() dir = false, want true")
	}
	if ignore.ShouldIgnore(filepath.Join(root, "logs-old", "app.log")) {
		t.Fatal("ShouldIgnore() sibling = true, want false")
	}
}

func TestIgnoreComponentMatchesRelativeEventPath(t *testing.T) {
	ignore := NewIgnoreComponent(".", []string{"logs/app.log"})

	if !ignore.ShouldIgnore(filepath.Join("logs", "app.log")) {
		t.Fatal("ShouldIgnore() relative = false, want true")
	}
}

func TestIgnoreComponentCleansPaths(t *testing.T) {
	root := filepath.Join(t.TempDir(), "watch")
	ignore := NewIgnoreComponent(root, []string{"./tmp/../logs/"})

	if !ignore.ShouldIgnore(filepath.Join(root, "logs", "app.log")) {
		t.Fatal("ShouldIgnore() cleaned = false, want true")
	}
}

func TestIgnoreComponentDoesNotMatchOutsideRoot(t *testing.T) {
	tempDir := t.TempDir()
	root := filepath.Join(tempDir, "watch")
	ignore := NewIgnoreComponent(root, []string{"file.txt"})

	if ignore.ShouldIgnore(filepath.Join(tempDir, "file.txt")) {
		t.Fatal("ShouldIgnore() outside root = true, want false")
	}
}
