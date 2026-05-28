//go:build !linux

package component

import (
	"errors"
	"os"

	"binmonitor/internal/types"
)

var errReadWatcherUnsupported = errors.New("read event monitoring is only supported on linux")

// ReadWatcherComponent 是非 Linux 平台的读取监听占位组件。
type ReadWatcherComponent struct{}

// NewReadWatcherComponent 在非 Linux 平台返回不支持错误。
func NewReadWatcherComponent() (*ReadWatcherComponent, error) {
	return nil, errReadWatcherUnsupported
}

// AddRecursiveWithFilter 在非 Linux 平台返回不支持错误。
func (w *ReadWatcherComponent) AddRecursiveWithFilter(root string, shouldSkip func(path string, info os.FileInfo) bool) error {
	return errReadWatcherUnsupported
}

// AddPath 在非 Linux 平台返回不支持错误。
func (w *ReadWatcherComponent) AddPath(path string) error {
	return errReadWatcherUnsupported
}

// RemovePath 在非 Linux 平台返回不支持错误。
func (w *ReadWatcherComponent) RemovePath(path string) error {
	return errReadWatcherUnsupported
}

// Events 返回空事件通道。
func (w *ReadWatcherComponent) Events() <-chan types.FileEvent {
	return nil
}

// Errors 返回空错误通道。
func (w *ReadWatcherComponent) Errors() <-chan error {
	return nil
}

// Close 在非 Linux 平台无操作。
func (w *ReadWatcherComponent) Close() error {
	return nil
}

// Run 在非 Linux 平台无操作。
func (w *ReadWatcherComponent) Run() {}
