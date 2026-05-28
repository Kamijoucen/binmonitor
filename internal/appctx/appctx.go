package appctx

import "binmonitor/internal/component"

// AppCtx 是持有各组件的应用上下文单例。
type AppCtx struct {
	watcher     *component.WatcherComponent
	state       *component.StateComponent
	ignore      *component.IgnoreComponent
	eventFilter *component.EventFilterComponent
	readWatcher *component.ReadWatcherComponent
}

// NewAppCtx 使用给定的组件创建一个 AppCtx。
func NewAppCtx(watcher *component.WatcherComponent, state *component.StateComponent, ignore *component.IgnoreComponent, eventFilter *component.EventFilterComponent, readWatcher *component.ReadWatcherComponent) *AppCtx {
	return &AppCtx{
		watcher:     watcher,
		state:       state,
		ignore:      ignore,
		eventFilter: eventFilter,
		readWatcher: readWatcher,
	}
}

// Watcher 返回 WatcherComponent。
func (a *AppCtx) Watcher() *component.WatcherComponent {
	return a.watcher
}

// State 返回 StateComponent。
func (a *AppCtx) State() *component.StateComponent {
	return a.state
}

// Ignore 返回 IgnoreComponent。
func (a *AppCtx) Ignore() *component.IgnoreComponent {
	return a.ignore
}

// EventFilter 返回 EventFilterComponent。
func (a *AppCtx) EventFilter() *component.EventFilterComponent {
	return a.eventFilter
}

// ReadWatcher 返回 ReadWatcherComponent。
func (a *AppCtx) ReadWatcher() *component.ReadWatcherComponent {
	return a.readWatcher
}
