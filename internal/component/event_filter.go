package component

import (
	"fmt"
	"strings"

	"binmonitor/internal/types"
)

// EventFilterComponent 管理需要输出的文件事件类型。
type EventFilterComponent struct {
	watching map[types.FileOp]struct{}
}

// NewEventFilterComponent 创建 EventFilterComponent。
func NewEventFilterComponent(events []string) (*EventFilterComponent, error) {
	filter := &EventFilterComponent{
		watching: make(map[types.FileOp]struct{}),
	}
	for _, event := range events {
		if err := filter.add(event); err != nil {
			return nil, err
		}
	}
	return filter, nil
}

// ShouldWatch 报告指定事件类型是否需要输出。
func (filter *EventFilterComponent) ShouldWatch(op types.FileOp) bool {
	if filter == nil {
		return true
	}
	_, ok := filter.watching[op]
	return ok
}

func (filter *EventFilterComponent) add(event string) error {
	switch strings.ToLower(strings.TrimSpace(event)) {
	case "create":
		filter.watching[types.OpCreate] = struct{}{}
	case "write", "modify":
		filter.watching[types.OpWrite] = struct{}{}
	case "remove":
		filter.watching[types.OpRemove] = struct{}{}
	case "delete":
		filter.watching[types.OpRemove] = struct{}{}
		filter.watching[types.OpRename] = struct{}{}
	case "rename":
		filter.watching[types.OpRename] = struct{}{}
	case "read":
		filter.watching[types.OpRead] = struct{}{}
	case "open", "process_open":
		filter.watching[types.OpOpen] = struct{}{}
	case "close", "process_close":
		filter.watching[types.OpClose] = struct{}{}
	case "":
		return nil
	default:
		return fmt.Errorf("unknown event type: %s", event)
	}
	return nil
}
