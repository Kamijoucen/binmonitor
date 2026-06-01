package main

import (
	"os"
	"path/filepath"
	"testing"

	"binmonitor/internal/types"
)

func TestParsePIDs(t *testing.T) {
	pids, err := parsePIDs("123, 456")
	if err != nil {
		t.Fatalf("parsePIDs() error = %v", err)
	}
	if len(pids) != 2 || pids[0] != 123 || pids[1] != 456 {
		t.Fatalf("pids = %#v, want 123 and 456", pids)
	}
}

func TestParsePIDsRejectsDuplicate(t *testing.T) {
	if _, err := parsePIDs("123,123"); err == nil {
		t.Fatal("parsePIDs() error = nil, want error")
	}
}

func TestLoadConfigPIDFlagUsesProcessDefaults(t *testing.T) {
	config, err := loadConfig([]string{"-pid", "123,456"})
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if config.Mode != types.ModeProcess {
		t.Fatalf("Mode = %q, want %q", config.Mode, types.ModeProcess)
	}
	if len(config.Events) != 2 || config.Events[0] != "open" || config.Events[1] != "close" {
		t.Fatalf("Events = %#v, want open and close", config.Events)
	}
	if len(config.Processes) != 2 || config.Processes[0].PID != 123 || config.Processes[1].PID != 456 {
		t.Fatalf("Processes = %#v, want pid 123 and 456", config.Processes)
	}
	if config.Processes[0].PollIntervalMs != types.DefaultProcessPollIntervalMs {
		t.Fatalf("PollIntervalMs = %d, want %d", config.Processes[0].PollIntervalMs, types.DefaultProcessPollIntervalMs)
	}
}

func TestLoadConfigPIDFlagOverridesConfiguredProcesses(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "binmonitor.json")
	data := []byte(`{"mode":"process","events":["open"],"processes":[{"pid":111,"name":"old"}]}`)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	config, err := loadConfig([]string{"-config", configPath, "-pid", "222,333", "-poll-ms", "50"})
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if len(config.Events) != 1 || config.Events[0] != "open" {
		t.Fatalf("Events = %#v, want configured open", config.Events)
	}
	if len(config.Processes) != 2 || config.Processes[0].PID != 222 || config.Processes[1].PID != 333 {
		t.Fatalf("Processes = %#v, want pid 222 and 333", config.Processes)
	}
	if config.Processes[0].PollIntervalMs != 50 || config.Processes[1].PollIntervalMs != 50 {
		t.Fatalf("Processes = %#v, want poll interval 50", config.Processes)
	}
}

func TestLoadConfigRejectsDirectoryWithPIDFlag(t *testing.T) {
	if _, err := loadConfig([]string{"-pid", "123", "/tmp"}); err == nil {
		t.Fatal("loadConfig() error = nil, want error")
	}
}
