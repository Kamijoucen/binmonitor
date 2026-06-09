//go:build !linux

package component

import (
	"fmt"

	"binmonitor/internal/types"
)

// NetWatcherComponent 非 Linux 平台的网络监控组件存根。
type NetWatcherComponent struct{}

// NewNetWatcherComponent 在非 Linux 平台上返回错误。
func NewNetWatcherComponent(config types.NetMonitorConfig) (*NetWatcherComponent, error) {
	return nil, fmt.Errorf("net watcher is only supported on Linux")
}

// Events 返回 nil channel。
func (w *NetWatcherComponent) Events() <-chan types.NetEvent { return nil }

// Errors 返回 nil channel。
func (w *NetWatcherComponent) Errors() <-chan error { return nil }

// Close 是无操作的。
func (w *NetWatcherComponent) Close() error { return nil }

// Run 是无操作的。
func (w *NetWatcherComponent) Run() {}
