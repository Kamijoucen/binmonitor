package logic

import (
	"fmt"
	"time"
)

// LogLineWriter 定义写入一行日志的能力。由 Component 实现。
type LogLineWriter interface {
	WriteLine(line string) error
}

// LogEvent 将事件记录格式化为日志行并写入日志文件（如果已启用）。
func LogEvent(lw LogLineWriter, record *EventRecord) error {
	if lw == nil {
		return nil
	}
	line := FormatEventRecord(record, time.Now().Format("2006-01-02 15:04:05"))
	return lw.WriteLine(line)
}

// FormatEventRecord 将事件记录格式化为控制台和文件日志共用的文本行。
func FormatEventRecord(record *EventRecord, timestamp string) string {
	if record.HasProcessMeta {
		namePart := ""
		if record.ProcessName != "" {
			namePart = fmt.Sprintf(" name=%s", record.ProcessName)
		}
		// 网络事件：不显示文件大小
		if isNetworkOp(record.Op) {
			return fmt.Sprintf("%s %s pid=%d fd=%d%s %s",
				timestamp, record.Op, record.PID, record.FD, namePart, record.Path,
			)
		}
		return fmt.Sprintf("%s %s pid=%d fd=%d%s %s %s",
			timestamp, record.Op, record.PID, record.FD, namePart, record.Path,
			HumanReadableSize(record.NewSize),
		)
	}

	diff := record.NewSize - record.OldSize
	return fmt.Sprintf("%s %s %s %s → %s (%s)",
		timestamp, record.Op, record.Path,
		HumanReadableSize(record.OldSize),
		HumanReadableSize(record.NewSize),
		HumanReadableSize(diff),
	)
}

// isNetworkOp 判断操作类型是否为网络事件。
func isNetworkOp(op string) bool {
	switch op {
	case "TCP_CONNECT", "TCP_CLOSE", "TCP_STATE", "UDP_CONNECT", "DNS_QUERY", "SOCKS5_CONNECT":
		return true
	}
	return false
}

// DedupRecordPath 返回用于去重日志的路径键。
func DedupRecordPath(record *EventRecord) string {
	if record.HasProcessMeta {
		if record.ProcessName != "" {
			return fmt.Sprintf("pid=%d name=%s %s", record.PID, record.ProcessName, record.Path)
		}
		return fmt.Sprintf("pid=%d %s", record.PID, record.Path)
	}
	return record.Path
}

// LogError 将错误信息写入日志文件（如果已启用）。
func LogError(lw LogLineWriter, format string, args ...interface{}) error {
	if lw == nil {
		return nil
	}
	line := fmt.Sprintf(format, args...)
	return lw.WriteLine(line)
}
