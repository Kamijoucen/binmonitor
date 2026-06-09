package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"binmonitor/internal/appctx"
	"binmonitor/internal/logic"
	"binmonitor/internal/types"
)

// MonitorService 是驱动文件监控的活跃逻辑入口。
type MonitorService struct {
	appCtx *appctx.AppCtx
	root   string
	mode   string
}

// NewMonitorService 创建一个 MonitorService。
func NewMonitorService(appCtx *appctx.AppCtx, config types.Config) *MonitorService {
	return &MonitorService{
		appCtx: appCtx,
		root:   config.Root,
		mode:   config.Mode,
	}
}

// Start 初始化状态，启动监听器，并阻塞直到被中断。
func (s *MonitorService) Start() error {
	if s.mode == types.ModeProcess {
		return s.startProcess()
	}
	if s.mode == types.ModeNetwork {
		return s.startNetwork()
	}
	return s.startDirectory()
}

func (s *MonitorService) startDirectory() error {
	if err := logic.InitStateFromPath(s.appCtx.State, s.appCtx.Ignore, s.root); err != nil {
		return fmt.Errorf("init state: %w", err)
	}
	if err := s.appCtx.Watcher.AddRecursiveWithFilter(s.root, func(path string, info os.FileInfo) bool {
		return info.IsDir() && s.appCtx.Ignore.ShouldIgnore(path)
	}); err != nil {
		return fmt.Errorf("add recursive watch: %w", err)
	}
	if s.appCtx.ReadWatcher != nil {
		if err := s.appCtx.ReadWatcher.AddRecursiveWithFilter(s.root, func(path string, info os.FileInfo) bool {
			return info.IsDir() && s.appCtx.Ignore.ShouldIgnore(path)
		}); err != nil {
			return fmt.Errorf("add recursive read watch: %w", err)
		}
	}

	go s.appCtx.Watcher.Run()
	var readEvents <-chan types.FileEvent
	var readErrors <-chan error
	if s.appCtx.ReadWatcher != nil {
		go s.appCtx.ReadWatcher.Run()
		readEvents = s.appCtx.ReadWatcher.Events()
		readErrors = s.appCtx.ReadWatcher.Errors()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case ev := <-s.appCtx.Watcher.Events():
			fe := logic.TranslateEvent(ev)
			if fe == nil {
				continue
			}
			record := logic.ProcessEvent(s.appCtx.State, s.appCtx.Ignore, s.appCtx.Watcher, s.appCtx.EventFilter, s.appCtx.ReadWatcher, *fe)
			if record != nil {
				s.printRecord(record)
			}
		case ev := <-readEvents:
			record := logic.ProcessEvent(s.appCtx.State, s.appCtx.Ignore, s.appCtx.Watcher, s.appCtx.EventFilter, s.appCtx.ReadWatcher, ev)
			if record != nil {
				s.printRecord(record)
			}
		case err := <-s.appCtx.Watcher.Errors():
			if err != nil {
				fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
				_ = logic.LogError(s.appCtx.LogWriter, "watcher error: %v", err)
			}
		case err := <-readErrors:
			if err != nil {
				fmt.Fprintf(os.Stderr, "read watcher error: %v\n", err)
				_ = logic.LogError(s.appCtx.LogWriter, "read watcher error: %v", err)
			}
		case <-sigCh:
			_ = s.appCtx.Watcher.Close()
			if s.appCtx.ReadWatcher != nil {
				_ = s.appCtx.ReadWatcher.Close()
			}
			return nil
		}
	}
}

func (s *MonitorService) startProcess() error {
	if s.appCtx.ProcessWatcher == nil {
		return fmt.Errorf("process watcher is not configured")
	}

	go s.appCtx.ProcessWatcher.Run()
	processEvents := s.appCtx.ProcessWatcher.Events()
	processErrors := s.appCtx.ProcessWatcher.Errors()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for processEvents != nil || processErrors != nil {
		select {
		case ev, ok := <-processEvents:
			if !ok {
				processEvents = nil
				continue
			}
			record := logic.ProcessEvent(s.appCtx.State, s.appCtx.Ignore, nil, s.appCtx.EventFilter, nil, ev)
			if record != nil {
				s.printRecord(record)
			}
		case err, ok := <-processErrors:
			if !ok {
				processErrors = nil
				continue
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "process watcher error: %v\n", err)
				_ = logic.LogError(s.appCtx.LogWriter, "process watcher error: %v", err)
			}
		case <-sigCh:
			_ = s.appCtx.ProcessWatcher.Close()
			return nil
		}
	}
	return nil
}

func (s *MonitorService) startNetwork() error {
	if s.appCtx.NetWatcher == nil {
		return fmt.Errorf("net watcher is not configured")
	}

	go s.appCtx.NetWatcher.Run()
	netEvents := s.appCtx.NetWatcher.Events()
	netErrors := s.appCtx.NetWatcher.Errors()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for netEvents != nil || netErrors != nil {
		select {
		case ev, ok := <-netEvents:
			if !ok {
				netEvents = nil
				continue
			}
			record := logic.ProcessNetEvent(s.appCtx.EventFilter, ev)
			if record != nil {
				s.printRecord(record)
			}
		case err, ok := <-netErrors:
			if !ok {
				netErrors = nil
				continue
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "net watcher error: %v\n", err)
				_ = logic.LogError(s.appCtx.LogWriter, "net watcher error: %v", err)
			}
		case <-sigCh:
			_ = s.appCtx.NetWatcher.Close()
			return nil
		}
	}
	return nil
}

func (s *MonitorService) printRecord(record *logic.EventRecord) {
	fmt.Println(logic.FormatEventRecord(record, time.Now().Format("2006-01-02 15:04:05")))
	_ = logic.LogEvent(s.appCtx.LogWriter, record)

	if s.appCtx.DedupStats != nil {
		if s.appCtx.DedupStats.Add(record.Op, logic.DedupRecordPath(record)) {
			_ = os.WriteFile("binmonitor.dedup.log", []byte(s.appCtx.DedupStats.Format()), 0644)
		}
	}
}
