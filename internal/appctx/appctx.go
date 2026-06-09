package appctx

import "binmonitor/internal/component"

// AppCtx 是持有各组件的应用上下文单例——装配根与生命周期管理。
type AppCtx struct {
	Watcher        *component.WatcherComponent
	State          *component.StateComponent
	Ignore         *component.IgnoreComponent
	EventFilter    *component.EventFilterComponent
	ReadWatcher    *component.ReadWatcherComponent
	ProcessWatcher *component.MultiProcessWatcherComponent
	NetWatcher     *component.NetWatcherComponent
	LogWriter      *component.LogWriterComponent
	DedupStats     *component.DedupStatsComponent
}

// NewAppCtx 使用给定的组件创建一个 AppCtx。
func NewAppCtx(watcher *component.WatcherComponent, state *component.StateComponent, ignore *component.IgnoreComponent, eventFilter *component.EventFilterComponent, readWatcher *component.ReadWatcherComponent, processWatcher *component.MultiProcessWatcherComponent, netWatcher *component.NetWatcherComponent, logWriter *component.LogWriterComponent, dedupStats *component.DedupStatsComponent) *AppCtx {
	return &AppCtx{
		Watcher:        watcher,
		State:          state,
		Ignore:         ignore,
		EventFilter:    eventFilter,
		ReadWatcher:    readWatcher,
		ProcessWatcher: processWatcher,
		NetWatcher:     netWatcher,
		LogWriter:      logWriter,
		DedupStats:     dedupStats,
	}
}
