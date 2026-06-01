package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

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

	if config.Log {
		config.Ignore = append(config.Ignore, "binmonitor.log")
	}
	if config.DedupLog {
		config.Ignore = append(config.Ignore, "binmonitor.dedup.log")
	}

	var watcherComp *component.WatcherComponent
	if config.Mode == types.ModeDirectory {
		info, err := os.Stat(config.Root)
		if err != nil || !info.IsDir() {
			fmt.Fprintf(os.Stderr, "invalid directory: %s\n", config.Root)
			os.Exit(1)
		}
		watcherComp, err = component.NewWatcherComponent()
		if err != nil {
			fmt.Fprintf(os.Stderr, "create watcher: %v\n", err)
			os.Exit(1)
		}
	}
	stateComp := component.NewStateComponent()
	ignoreComp := component.NewIgnoreComponent(config.Root, config.Ignore)
	eventFilterComp, err := component.NewEventFilterComponent(config.Events)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create event filter: %v\n", err)
		os.Exit(1)
	}
	var readWatcherComp *component.ReadWatcherComponent
	if config.Mode == types.ModeDirectory && eventFilterComp.ShouldWatch(types.OpRead) {
		readWatcherComp, err = component.NewReadWatcherComponent()
		if err != nil {
			fmt.Fprintf(os.Stderr, "create read watcher: %v\n", err)
			os.Exit(1)
		}
	}
	var processWatcherComp *component.MultiProcessWatcherComponent
	if config.Mode == types.ModeProcess {
		processWatcherComp, err = component.NewMultiProcessWatcherComponent(config.Processes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create process watcher: %v\n", err)
			os.Exit(1)
		}
	}
	var logWriterComp *component.LogWriterComponent
	if config.Log {
		logWriterComp, err = component.NewLogWriterComponent("binmonitor.log")
		if err != nil {
			fmt.Fprintf(os.Stderr, "create log writer: %v\n", err)
			os.Exit(1)
		}
		defer logWriterComp.Close()
	}
	var dedupStatsComp *component.DedupStatsComponent
	if config.DedupLog {
		dedupStatsComp = component.NewDedupStatsComponent()
	}
	appCtx := appctx.NewAppCtx(watcherComp, stateComp, ignoreComp, eventFilterComp, readWatcherComp, processWatcherComp, logWriterComp, dedupStatsComp)

	monitorSvc := service.NewMonitorService(appCtx, config)
	if err := monitorSvc.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "monitor service: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig(args []string) (types.Config, error) {
	flagSet := flag.NewFlagSet("binmonitor", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)
	configPath := flagSet.String("config", defaultConfigPath, "config file path")
	pidValue := flagSet.String("pid", "", "comma-separated process IDs to monitor")
	pollMS := flagSet.Int("poll-ms", 0, "process fd polling interval in milliseconds")
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
		if strings.TrimSpace(*pidValue) != "" {
			return types.Config{}, fmt.Errorf("unexpected monitor directory with -pid: %s", remainingArgs[0])
		}
		if configPathSet {
			return types.Config{}, fmt.Errorf("unexpected monitor directory with -config: %s", remainingArgs[0])
		}
		config.Root = remainingArgs[0]
	}
	if *pollMS < 0 {
		return types.Config{}, fmt.Errorf("-poll-ms must be greater than or equal to 0")
	}
	if strings.TrimSpace(*pidValue) != "" {
		pids, err := parsePIDs(*pidValue)
		if err != nil {
			return types.Config{}, err
		}
		config.Mode = types.ModeProcess
		config.Process = nil
		config.ProcessPollIntervalMs = *pollMS
		config.Processes = make([]types.ProcessConfig, 0, len(pids))
		for _, pid := range pids {
			config.Processes = append(config.Processes, types.ProcessConfig{PID: pid, PollIntervalMs: *pollMS})
		}
	}
	if err := logic.NormalizeConfig(&config); err != nil {
		return types.Config{}, err
	}
	return config, nil
}

func parsePIDs(value string) ([]int, error) {
	parts := strings.Split(value, ",")
	pids := make([]int, 0, len(parts))
	seen := make(map[int]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		pid, err := strconv.Atoi(part)
		if err != nil || pid <= 0 {
			return nil, fmt.Errorf("invalid pid: %s", part)
		}
		if _, ok := seen[pid]; ok {
			return nil, fmt.Errorf("duplicate pid: %d", pid)
		}
		seen[pid] = struct{}{}
		pids = append(pids, pid)
	}
	if len(pids) == 0 {
		return nil, fmt.Errorf("-pid requires at least one pid")
	}
	return pids, nil
}
