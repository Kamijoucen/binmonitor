package component

import (
	"container/list"
	"strings"
	"sync"
)

// DedupStatsComponent 记录去重的事件统计。
// 每个文件的每种事件类型在当前进程生命周期内仅记录一次；当同一事件再次发生时，
// 该记录会像冒泡一样移动到对应事件组的最末尾，保证中间其他文件的相对时序不变。
type DedupStatsComponent struct {
	mu     sync.Mutex
	groups map[string]*list.List
	pos    map[string]*list.Element
}

// NewDedupStatsComponent 创建一个新的 DedupStatsComponent。
func NewDedupStatsComponent() *DedupStatsComponent {
	return &DedupStatsComponent{
		groups: make(map[string]*list.List),
		pos:    make(map[string]*list.Element),
	}
}

// Add 尝试添加一个事件记录。
// 若该 (Op, Path) 组合已存在，则将其移到对应事件组的最末尾并返回 true；
// 否则加入并返回 true。调用方每次返回 true 时都可以触发写文件。
func (d *DedupStatsComponent) Add(op string, path string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := op + ":" + path
	if elem, ok := d.pos[key]; ok {
		d.groups[op].MoveToBack(elem)
		return true
	}

	if d.groups[op] == nil {
		d.groups[op] = list.New()
	}
	elem := d.groups[op].PushBack(path)
	d.pos[key] = elem
	return true
}

// Format 将当前统计按事件类型分组格式化为纯文本字符串。
// 组内按最近一次发生顺序排列（越晚发生越靠下），中间元素的相对时序保持不变。
func (d *DedupStatsComponent) Format() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	var b strings.Builder
	firstGroup := true
	for _, op := range []string{"CREATE", "WRITE", "REMOVE", "RENAME", "READ", "CHMOD"} {
		grp, ok := d.groups[op]
		if !ok || grp.Len() == 0 {
			continue
		}
		if !firstGroup {
			b.WriteString("\n")
		}
		firstGroup = false
		b.WriteString("[")
		b.WriteString(op)
		b.WriteString("]\n")
		for elem := grp.Front(); elem != nil; elem = elem.Next() {
			b.WriteString(elem.Value.(string))
			b.WriteString("\n")
		}
	}
	return b.String()
}
