//go:build linux

package component

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"binmonitor/internal/logic"
	"binmonitor/internal/types"
)

// NetWatcherComponent 通过 /proc 轮询监听指定进程的网络连接变化。
// 遵循 Atomic Architecture：持有生命周期状态，被 AppCtx 持有和管理。
type NetWatcherComponent struct {
	config    types.NetMonitorConfig
	interval  time.Duration
	events    chan types.NetEvent
	errors    chan error
	done      chan struct{}
	closeOnce sync.Once
}

// NewNetWatcherComponent 创建网络连接监听组件。
func NewNetWatcherComponent(config types.NetMonitorConfig) (*NetWatcherComponent, error) {
	if config.PID <= 0 {
		return nil, fmt.Errorf("net watcher: pid must be greater than 0")
	}
	if config.PollIntervalMs <= 0 {
		config.PollIntervalMs = types.DefaultNetworkPollIntervalMs
	}
	if config.Socks5Ports == nil {
		config.Socks5Ports = types.DefaultSocks5Ports
	}
	return &NetWatcherComponent{
		config:   config,
		interval: time.Duration(config.PollIntervalMs) * time.Millisecond,
		events:   make(chan types.NetEvent, 128),
		errors:   make(chan error, 64),
		done:     make(chan struct{}),
	}, nil
}

// Events 返回网络事件通道。
func (w *NetWatcherComponent) Events() <-chan types.NetEvent {
	return w.events
}

// Errors 返回错误通道。
func (w *NetWatcherComponent) Errors() <-chan error {
	return w.errors
}

// Close 停止网络监听。
func (w *NetWatcherComponent) Close() error {
	w.closeOnce.Do(func() {
		close(w.done)
	})
	return nil
}

// Run 开始轮询目标进程的网络连接快照。
func (w *NetWatcherComponent) Run() {
	defer close(w.events)
	defer close(w.errors)

	previous, err := logic.ReadProcNetFiles(w.config.PID)
	if err != nil {
		w.sendError(fmt.Errorf("net watcher initial snapshot pid=%d: %w", w.config.PID, err))
		return
	}
	if previous == nil {
		previous = make(map[uint64]types.NetConnInfo)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-ticker.C:
			next, err := logic.ReadProcNetFiles(w.config.PID)
			if err != nil {
				w.sendError(fmt.Errorf("net watcher snapshot pid=%d: %w", w.config.PID, err))
				return
			}
			if next == nil {
				next = make(map[uint64]types.NetConnInfo)
			}
			for _, event := range diffNetSnapshots(w.config, previous, next) {
				if !w.sendEvent(event) {
					return
				}
			}
			previous = next
		}
	}
}

func (w *NetWatcherComponent) sendEvent(event types.NetEvent) bool {
	select {
	case w.events <- event:
		return true
	case <-w.done:
		return false
	}
}

func (w *NetWatcherComponent) sendError(err error) bool {
	select {
	case w.errors <- err:
		return true
	case <-w.done:
		return false
	}
}

// diffNetSnapshots 比较两次网络快照，生成连接建立/关闭事件。
func diffNetSnapshots(config types.NetMonitorConfig, previous map[uint64]types.NetConnInfo, next map[uint64]types.NetConnInfo) []types.NetEvent {
	// 收集所有 inode
	inodes := make(map[uint64]struct{}, len(previous)+len(next))
	for inode := range previous {
		inodes[inode] = struct{}{}
	}
	for inode := range next {
		inodes[inode] = struct{}{}
	}

	orderedInodes := make([]uint64, 0, len(inodes))
	for inode := range inodes {
		orderedInodes = append(orderedInodes, inode)
	}
	sort.Slice(orderedInodes, func(i, j int) bool { return orderedInodes[i] < orderedInodes[j] })

	events := make([]types.NetEvent, 0)
	socks5PortSet := logic.BuildPortSet(config.Socks5Ports)

	for _, inode := range orderedInodes {
		prev, hadPrev := previous[inode]
		next, hasNext := next[inode]

		// 新连接建立
		if hasNext && !hadPrev {
			op := logic.ClassifyNetOp(next, config.DNSTrace, socks5PortSet)
			events = append(events, netConnInfoToEvent(next, op))
			continue
		}

		// 连接关闭（之前有，现在没有了）
		if hadPrev && !hasNext {
			events = append(events, netConnInfoToEvent(prev, types.OpTCPClose))
			continue
		}

		// 连接状态变化（如 ESTABLISHED → CLOSE_WAIT）
		if hadPrev && hasNext && prev.State != next.State && prev.Protocol == next.Protocol &&
			strings.HasPrefix(next.Protocol, "tcp") {
			events = append(events, netConnInfoToEvent(next, types.OpTCPStateChange))
		}
	}

	return events
}

// netConnInfoToEvent 将内部连接信息转换为 NetEvent。
func netConnInfoToEvent(info types.NetConnInfo, op types.NetOp) types.NetEvent {
	return types.NetEvent{
		Protocol: info.Protocol,
		SrcIP:    info.SrcIP,
		SrcPort:  info.SrcPort,
		DstIP:    info.DstIP,
		DstPort:  info.DstPort,
		Op:       op,
	}
}
