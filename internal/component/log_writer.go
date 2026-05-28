package component

import (
	"fmt"
	"os"
	"sync"
)

// LogWriterComponent 持有日志文件状态，提供线程安全的行写入能力。
type LogWriterComponent struct {
	mu   sync.Mutex
	file *os.File
}

// NewLogWriterComponent 以追加模式打开指定路径的日志文件。
func NewLogWriterComponent(path string) (*LogWriterComponent, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	return &LogWriterComponent{file: file}, nil
}

// WriteLine 线程安全地写入一行内容并追加换行符。
func (l *LogWriterComponent) WriteLine(line string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err := fmt.Fprintln(l.file, line)
	return err
}

// Close 关闭日志文件。
func (l *LogWriterComponent) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
