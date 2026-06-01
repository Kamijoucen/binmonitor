package logic

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"binmonitor/internal/types"
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

func TestNormalizeConfigDefaultsToDirectoryMode(t *testing.T) {
	config := types.Config{}
	if err := NormalizeConfig(&config); err != nil {
		t.Fatalf("NormalizeConfig() error = %v", err)
	}
	if config.Mode != types.ModeDirectory {
		t.Fatalf("Mode = %q, want %q", config.Mode, types.ModeDirectory)
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

func TestLoadConfigLoadsProcesses(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "binmonitor.json")
	data := []byte(`{"mode":"process","processPollIntervalMs":300,"processes":[{"pid":123,"name":"api"},{"pid":456,"name":"worker","pollIntervalMs":100}]}`)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.Mode != types.ModeProcess {
		t.Fatalf("Mode = %q, want %q", config.Mode, types.ModeProcess)
	}
	if len(config.Events) != 2 || config.Events[0] != "open" || config.Events[1] != "close" {
		t.Fatalf("Events = %#v, want open and close", config.Events)
	}
	if len(config.Processes) != 2 {
		t.Fatalf("Processes length = %d, want 2", len(config.Processes))
	}
	if config.Processes[0].PID != 123 || config.Processes[0].Name != "api" || config.Processes[0].PollIntervalMs != 300 {
		t.Fatalf("Processes[0] = %#v, want pid 123 name api interval 300", config.Processes[0])
	}
	if config.Processes[1].PID != 456 || config.Processes[1].Name != "worker" || config.Processes[1].PollIntervalMs != 100 {
		t.Fatalf("Processes[1] = %#v, want pid 456 name worker interval 100", config.Processes[1])
	}
}

func TestLoadConfigInfersProcessModeFromProcesses(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "binmonitor.json")
	if err := os.WriteFile(configPath, []byte(`{"processes":[{"pid":123}]}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.Mode != types.ModeProcess {
		t.Fatalf("Mode = %q, want %q", config.Mode, types.ModeProcess)
	}
	if config.Processes[0].PollIntervalMs != types.DefaultProcessPollIntervalMs {
		t.Fatalf("PollIntervalMs = %d, want %d", config.Processes[0].PollIntervalMs, types.DefaultProcessPollIntervalMs)
	}
}

func TestLoadConfigRejectsProcessModeWithoutPID(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "binmonitor.json")
	if err := os.WriteFile(configPath, []byte(`{"mode":"process"}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := LoadConfig(configPath); err == nil {
		t.Fatal("LoadConfig() error = nil, want error")
	}
}

func TestLoadConfigRejectsDuplicateProcessPID(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "binmonitor.json")
	if err := os.WriteFile(configPath, []byte(`{"mode":"process","processes":[{"pid":123},{"pid":123}]}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := LoadConfig(configPath); err == nil {
		t.Fatal("LoadConfig() error = nil, want error")
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
