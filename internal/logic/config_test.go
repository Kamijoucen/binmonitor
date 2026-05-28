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
}

func TestLoadConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "binmonitor.json")
	data := []byte(`{"root":"/tmp/watch","ignore":["logs","tmp/cache.db"]}`)
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
