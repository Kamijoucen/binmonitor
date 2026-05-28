package logic

import (
	"fmt"
	"time"

	"binmonitor/internal/appctx"
)

// LogEvent 将事件记录格式化为日志行并写入日志文件（如果已启用）。
func LogEvent(appCtx *appctx.AppCtx, record *EventRecord) error {
	lw := appCtx.LogWriter()
	if lw == nil {
		return nil
	}
	diff := record.NewSize - record.OldSize
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("%s %s %s %s → %s (%s)",
		timestamp, record.Op, record.Path,
		HumanReadableSize(record.OldSize),
		HumanReadableSize(record.NewSize),
		HumanReadableSize(diff),
	)
	return lw.WriteLine(line)
}

// LogError 将错误信息写入日志文件（如果已启用）。
func LogError(appCtx *appctx.AppCtx, format string, args ...interface{}) error {
	lw := appCtx.LogWriter()
	if lw == nil {
		return nil
	}
	line := fmt.Sprintf(format, args...)
	return lw.WriteLine(line)
}
