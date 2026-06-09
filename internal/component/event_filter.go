package component

import "binmonitor/internal/types"

// EventFilterComponent 管理需要输出的文件事件类型。
type EventFilterComponent struct {
	watching map[types.FileOp]struct{}
}

// NewEventFilterComponent 创建 EventFilterComponent，接收已由 logic 层解析好的 FileOp 列表。
func NewEventFilterComponent(ops []types.FileOp) *EventFilterComponent {
	filter := &EventFilterComponent{
		watching: make(map[types.FileOp]struct{}),
	}
	for _, op := range ops {
		filter.watching[op] = struct{}{}
	}
	return filter
}

// ShouldWatch 报告指定事件类型是否需要输出。
func (filter *EventFilterComponent) ShouldWatch(op types.FileOp) bool {
	if filter == nil {
		return true
	}
	_, ok := filter.watching[op]
	return ok
}
