package logic

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.Root != "." {
		t.Fatalf("Root = %q, want %q", config.Root, ".")
	}
	if len(config.Ignore) != 0 {
		t.Fatalf("Ignore length = %d, want 0", len(config.Ignore))
	}
	if len(config.Events) != 4 {
		t.Fatalf("Events length = %d, want 4", len(config.Events))
	}
}

func TestLoadConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "binmonitor.json")
	data := []byte(`{"root":"/tmp/watch","ignore":["logs","tmp/cache.db"],"events":["create","read"]}`)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.Root != "/tmp/watch" {
		t.Fatalf("Root = %q, want %q", config.Root, "/tmp/watch")
	}
	if len(config.Ignore) != 2 || config.Ignore[0] != "logs" || config.Ignore[1] != "tmp/cache.db" {
		t.Fatalf("Ignore = %#v, want logs and tmp/cache.db", config.Ignore)
	}
	if len(config.Events) != 2 || config.Events[0] != "create" || config.Events[1] != "read" {
		t.Fatalf("Events = %#v, want create and read", config.Events)
	}
}

func TestLoadConfigUsesDefaults(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "binmonitor.json")
	if err := os.WriteFile(configPath, []byte(`{}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.Root != "." {
		t.Fatalf("Root = %q, want %q", config.Root, ".")
	}
	if len(config.Ignore) != 0 {
		t.Fatalf("Ignore length = %d, want 0", len(config.Ignore))
	}
	if len(config.Events) != 4 {
		t.Fatalf("Events length = %d, want 4", len(config.Events))
	}
}

func TestWriteDefaultConfigDoesNotOverwrite(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "binmonitor.json")
	if err := WriteDefaultConfig(configPath); err != nil {
		t.Fatalf("WriteDefaultConfig() error = %v", err)
	}

	if err := WriteDefaultConfig(configPath); !errors.Is(err, os.ErrExist) {
		t.Fatalf("WriteDefaultConfig() second error = %v, want os.ErrExist", err)
	}
}
