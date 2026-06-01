//go:build !linux

package component

import (
	"errors"

	"binmonitor/internal/types"
)

var errProcessWatcherUnsupported = errors.New("process file monitoring is only supported on linux")

// NewProcessWatcherComponent 在非 Linux 平台返回不支持错误。
func NewProcessWatcherComponent(process types.ProcessConfig) (*ProcessWatcherComponent, error) {
	return nil, errProcessWatcherUnsupported
}

// Run 在非 Linux 平台无操作。
func (w *ProcessWatcherComponent) Run() {}
