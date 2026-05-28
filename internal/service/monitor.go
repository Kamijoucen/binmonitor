package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"binmonitor/internal/appctx"
	"binmonitor/internal/logic"
)

// MonitorService is the active logic entry that drives file monitoring.
type MonitorService struct {
	appCtx *appctx.AppCtx
	root   string
}

// NewMonitorService creates a MonitorService.
func NewMonitorService(appCtx *appctx.AppCtx, root string) *MonitorService {
	return &MonitorService{
		appCtx: appCtx,
		root:   root,
	}
}

// Start initializes state, starts the watcher, and blocks until interrupted.
func (s *MonitorService) Start() error {
	if err := s.appCtx.State().InitFromPath(s.root); err != nil {
		return fmt.Errorf("init state: %w", err)
	}
	if err := s.appCtx.Watcher().AddRecursive(s.root); err != nil {
		return fmt.Errorf("add recursive watch: %w", err)
	}

	go s.appCtx.Watcher().Run()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case ev := <-s.appCtx.Watcher().Events():
			logic.ProcessEvent(s.appCtx, ev)
		case err := <-s.appCtx.Watcher().Errors():
			if err != nil {
				fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
			}
		case <-sigCh:
			_ = s.appCtx.Watcher().Close()
			return nil
		}
	}
}
