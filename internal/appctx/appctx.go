package appctx

import "binmonitor/internal/component"

// AppCtx is the application context singleton holding components.
type AppCtx struct {
	watcher *component.WatcherComponent
	state   *component.StateComponent
}

// NewAppCtx creates an AppCtx with the given components.
func NewAppCtx(watcher *component.WatcherComponent, state *component.StateComponent) *AppCtx {
	return &AppCtx{
		watcher: watcher,
		state:   state,
	}
}

// Watcher returns the WatcherComponent.
func (a *AppCtx) Watcher() *component.WatcherComponent {
	return a.watcher
}

// State returns the StateComponent.
func (a *AppCtx) State() *component.StateComponent {
	return a.state
}
