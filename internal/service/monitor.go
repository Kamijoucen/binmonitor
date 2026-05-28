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
}

// NewMonitorService 创建一个 MonitorService。
func NewMonitorService(appCtx *appctx.AppCtx, root string) *MonitorService {
	return &MonitorService{
		appCtx: appCtx,
		root:   root,
	}
}

// Start 初始化状态，启动监听器，并阻塞直到被中断。
func (s *MonitorService) Start() error {
	if err := logic.InitStateFromPath(s.appCtx, s.root); err != nil {
		return fmt.Errorf("init state: %w", err)
	}
	if err := s.appCtx.Watcher().AddRecursiveWithFilter(s.root, func(path string, info os.FileInfo) bool {
		return info.IsDir() && s.appCtx.Ignore().ShouldIgnore(path)
	}); err != nil {
		return fmt.Errorf("add recursive watch: %w", err)
	}
	if s.appCtx.ReadWatcher() != nil {
		if err := s.appCtx.ReadWatcher().AddRecursiveWithFilter(s.root, func(path string, info os.FileInfo) bool {
			return info.IsDir() && s.appCtx.Ignore().ShouldIgnore(path)
		}); err != nil {
			return fmt.Errorf("add recursive read watch: %w", err)
		}
	}

	go s.appCtx.Watcher().Run()
	var readEvents <-chan types.FileEvent
	var readErrors <-chan error
	if s.appCtx.ReadWatcher() != nil {
		go s.appCtx.ReadWatcher().Run()
		readEvents = s.appCtx.ReadWatcher().Events()
		readErrors = s.appCtx.ReadWatcher().Errors()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case ev := <-s.appCtx.Watcher().Events():
			fe := logic.TranslateEvent(ev)
			if fe == nil {
				continue
			}
			record := logic.ProcessEvent(s.appCtx, *fe)
			if record != nil {
				s.printRecord(record)
			}
		case ev := <-readEvents:
			record := logic.ProcessEvent(s.appCtx, ev)
			if record != nil {
				s.printRecord(record)
			}
		case err := <-s.appCtx.Watcher().Errors():
			if err != nil {
				fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
				_ = logic.LogError(s.appCtx, "watcher error: %v", err)
			}
		case err := <-readErrors:
			if err != nil {
				fmt.Fprintf(os.Stderr, "read watcher error: %v\n", err)
				_ = logic.LogError(s.appCtx, "read watcher error: %v", err)
			}
		case <-sigCh:
			_ = s.appCtx.Watcher().Close()
			if s.appCtx.ReadWatcher() != nil {
				_ = s.appCtx.ReadWatcher().Close()
			}
			return nil
		}
	}
}

func (s *MonitorService) printRecord(record *logic.EventRecord) {
	diff := record.NewSize - record.OldSize
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%s %s %s %s → %s (%s)\n",
		timestamp, record.Op, record.Path,
		logic.HumanReadableSize(record.OldSize),
		logic.HumanReadableSize(record.NewSize),
		logic.HumanReadableSize(diff),
	)
	_ = logic.LogEvent(s.appCtx, record)
}
