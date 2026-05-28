package logic

import (
	"encoding/json"
	"fmt"
	"os"

	"binmonitor/internal/types"
)

// DefaultConfig 返回默认启动配置。
func DefaultConfig() types.Config {
	return types.Config{
		Root:   ".",
		Ignore: []string{},
		Events: []string{"create", "write", "remove", "rename"},
	}
}

// LoadConfig 从 JSON 文件读取启动配置，缺省字段使用默认值。
func LoadConfig(path string) (types.Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("parse config: %w", err)
	}
	if config.Root == "" {
		config.Root = "."
	}
	if config.Ignore == nil {
		config.Ignore = []string{}
	}
	if config.Events == nil {
		config.Events = DefaultConfig().Events
	}
	return config, nil
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
