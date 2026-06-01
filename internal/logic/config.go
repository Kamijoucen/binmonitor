package logic

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"binmonitor/internal/types"
)

// DefaultConfig 返回默认启动配置。
func DefaultConfig() types.Config {
	return types.Config{
		Root:     ".",
		Ignore:   []string{},
		Events:   []string{"create", "write", "remove", "rename"},
		Log:      false,
		DedupLog: false,
	}
}

// LoadConfig 从 JSON 文件读取启动配置，缺省字段使用默认值。
func LoadConfig(path string) (types.Config, error) {
	config := types.Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	config.EventsConfigured = jsonHasTopLevelField(data, "events")
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("parse config: %w", err)
	}
	if err := NormalizeConfig(&config); err != nil {
		return config, err
	}
	return config, nil
}

// NormalizeConfig 补全配置默认值并校验进程监控配置。
func NormalizeConfig(config *types.Config) error {
	if config.Root == "" {
		config.Root = "."
	}
	if config.Ignore == nil {
		config.Ignore = []string{}
	}

	mode := strings.ToLower(strings.TrimSpace(config.Mode))
	if mode == "" {
		if config.Process != nil || len(config.Processes) > 0 {
			mode = types.ModeProcess
		} else {
			mode = types.ModeDirectory
		}
	}
	config.Mode = mode

	switch config.Mode {
	case types.ModeDirectory:
		if config.Events == nil {
			config.Events = DefaultConfig().Events
		}
	case types.ModeProcess:
		if config.Events == nil || !config.EventsConfigured {
			config.Events = []string{"open", "close"}
		}
		if err := normalizeProcessConfig(config); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown mode: %s", config.Mode)
	}
	return nil
}

func normalizeProcessConfig(config *types.Config) error {
	processes := append([]types.ProcessConfig{}, config.Processes...)
	if config.Process != nil {
		processes = append([]types.ProcessConfig{*config.Process}, processes...)
	}
	if len(processes) == 0 {
		return fmt.Errorf("process mode requires at least one process")
	}

	defaultInterval := config.ProcessPollIntervalMs
	if defaultInterval < 0 {
		return fmt.Errorf("process poll interval must be greater than or equal to 0")
	}
	if defaultInterval <= 0 {
		defaultInterval = types.DefaultProcessPollIntervalMs
	}
	config.ProcessPollIntervalMs = defaultInterval
	config.Process = nil

	seen := make(map[int]struct{}, len(processes))
	for idx := range processes {
		processes[idx].Name = strings.TrimSpace(processes[idx].Name)
		if processes[idx].PID <= 0 {
			return fmt.Errorf("invalid process pid at index %d: %d", idx, processes[idx].PID)
		}
		if _, ok := seen[processes[idx].PID]; ok {
			return fmt.Errorf("duplicate process pid: %d", processes[idx].PID)
		}
		seen[processes[idx].PID] = struct{}{}
		if processes[idx].PollIntervalMs < 0 {
			return fmt.Errorf("invalid process poll interval at index %d: %d", idx, processes[idx].PollIntervalMs)
		}
		if processes[idx].PollIntervalMs <= 0 {
			processes[idx].PollIntervalMs = defaultInterval
		}
	}
	config.Processes = processes
	return nil
}

func jsonHasTopLevelField(data []byte, field string) bool {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return false
	}
	_, ok := raw[field]
	return ok
}

// WriteDefaultConfig 将默认配置写入指定路径，不覆盖已存在文件。
func WriteDefaultConfig(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(DefaultConfig()); err != nil {
		return fmt.Errorf("write default config: %w", err)
	}
	return nil
}
