package main

import (
	"fmt"
	"os"

	"binmonitor/internal/appctx"
	"binmonitor/internal/component"
	"binmonitor/internal/infra"
	"binmonitor/internal/service"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "invalid directory: %s\n", root)
		os.Exit(1)
	}

	fw, err := infra.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "create watcher: %v\n", err)
		os.Exit(1)
	}

	watcherComp := component.NewWatcherComponent(fw)
	stateComp := component.NewStateComponent()
	appCtx := appctx.NewAppCtx(watcherComp, stateComp)

	monitorSvc := service.NewMonitorService(appCtx, root)
	if err := monitorSvc.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "monitor service: %v\n", err)
		os.Exit(1)
	}
}
