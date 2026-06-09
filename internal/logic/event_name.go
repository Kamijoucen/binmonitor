package logic

import (
	"fmt"
	"strings"

	"binmonitor/internal/types"
)

// ResolveEventOps 将事件名称解析为对应的 FileOp 列表。
// 支持别名：modify→write, delete→remove+rename。
func ResolveEventOps(eventName string) ([]types.FileOp, error) {
	switch strings.ToLower(strings.TrimSpace(eventName)) {
	case "create":
		return []types.FileOp{types.OpCreate}, nil
	case "write", "modify":
		return []types.FileOp{types.OpWrite}, nil
	case "remove":
		return []types.FileOp{types.OpRemove}, nil
	case "delete":
		return []types.FileOp{types.OpRemove, types.OpRename}, nil
	case "rename":
		return []types.FileOp{types.OpRename}, nil
	case "read":
		return []types.FileOp{types.OpRead}, nil
	case "open", "process_open":
		return []types.FileOp{types.OpOpen}, nil
	case "close", "process_close":
		return []types.FileOp{types.OpClose}, nil
	case "":
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventName)
	}
}

// ResolveAllEventOps 批量解析事件名称列表，去重后返回 FileOp 集合。
func ResolveAllEventOps(eventNames []string) ([]types.FileOp, error) {
	seen := make(map[types.FileOp]struct{})
	var ops []types.FileOp
	for _, name := range eventNames {
		resolved, err := ResolveEventOps(name)
		if err != nil {
			return nil, err
		}
		for _, op := range resolved {
			if _, ok := seen[op]; !ok {
				seen[op] = struct{}{}
				ops = append(ops, op)
			}
		}
	}
	return ops, nil
}
