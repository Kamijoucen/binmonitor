package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"binmonitor/internal/appctx"
	"binmonitor/internal/component"
	"binmonitor/internal/logic"
	"binmonitor/internal/service"
	"binmonitor/internal/types"
)

const defaultConfigPath = "binmonitor.json"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		if err := logic.WriteDefaultConfig(defaultConfigPath); err != nil {
			fmt.Fprintf(os.Stderr, "init config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("created %s\n", defaultConfigPath)
		return
	}

	config, err := loadConfig(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(config.Root)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "invalid directory: %s\n", config.Root)
		os.Exit(1)
	}

	watcherComp, err := component.NewWatcherComponent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "create watcher: %v\n", err)
		os.Exit(1)
	}
	stateComp := component.NewStateComponent()
	ignoreComp := component.NewIgnoreComponent(config.Root, config.Ignore)
	appCtx := appctx.NewAppCtx(watcherComp, stateComp, ignoreComp)

	monitorSvc := service.NewMonitorService(appCtx, config.Root)
	if err := monitorSvc.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "monitor service: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig(args []string) (types.Config, error) {
	flagSet := flag.NewFlagSet("binmonitor", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)
	configPath := flagSet.String("config", defaultConfigPath, "config file path")
	if err := flagSet.Parse(args); err != nil {
		return types.Config{}, err
	}

	configPathSet := false
	flagSet.Visit(func(flagValue *flag.Flag) {
		if flagValue.Name == "config" {
			configPathSet = true
		}
	})

	config, err := logic.LoadConfig(*configPath)
	if err != nil {
		if !configPathSet && errors.Is(err, os.ErrNotExist) {
			config = logic.DefaultConfig()
		} else {
			return types.Config{}, fmt.Errorf("load config %s: %w", *configPath, err)
		}
	}

	remainingArgs := flagSet.Args()
	if len(remainingArgs) > 1 {
		return types.Config{}, fmt.Errorf("unexpected arguments: %v", remainingArgs[1:])
	}
	if len(remainingArgs) == 1 {
		if configPathSet {
			return types.Config{}, fmt.Errorf("unexpected monitor directory with -config: %s", remainingArgs[0])
		}
		config.Root = remainingArgs[0]
	}
	return config, nil
}
