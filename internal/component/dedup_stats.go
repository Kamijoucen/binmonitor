package component

import (
	"strings"
	"sync"
)

// DedupStatsComponent 记录去重的事件统计。
// 每个文件的每种事件类型仅在当前进程生命周期内记录一次。
type DedupStatsComponent struct {
	mu     sync.Mutex
	seen   map[string]struct{}
	groups map[string][]string
}

// NewDedupStatsComponent 创建一个新的 DedupStatsComponent。
func NewDedupStatsComponent() *DedupStatsComponent {
	return &DedupStatsComponent{
		seen:   make(map[string]struct{}),
		groups: make(map[string][]string),
	}
}

// Add 尝试添加一个事件记录。
// 若该 (Op, Path) 组合已存在则返回 false；否则加入并返回 true。
func (d *DedupStatsComponent) Add(op string, path string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := op + ":" + path
	if _, ok := d.seen[key]; ok {
		return false
	}
	d.seen[key] = struct{}{}
	d.groups[op] = append(d.groups[op], path)
	return true
}

// Format 将当前统计按事件类型分组格式化为纯文本字符串。
// 组内按首次发生顺序排列（越早加入越靠上）。
func (d *DedupStatsComponent) Format() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	var b strings.Builder
	firstGroup := true
	for _, op := range []string{"CREATE", "WRITE", "REMOVE", "RENAME", "READ", "CHMOD"} {
		paths, ok := d.groups[op]
		if !ok || len(paths) == 0 {
			continue
		}
		if !firstGroup {
			b.WriteString("\n")
		}
		firstGroup = false
		b.WriteString("[")
		b.WriteString(op)
		b.WriteString("]\n")
		for _, path := range paths {
			b.WriteString(path)
			b.WriteString("\n")
		}
	}
	return b.String()
}
