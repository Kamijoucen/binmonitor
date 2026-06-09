package component

import (
	"testing"

	"binmonitor/internal/types"
)

func TestEventFilterComponentShouldWatchConfiguredEvents(t *testing.T) {
	ops := []types.FileOp{types.OpCreate, types.OpWrite, types.OpRemove, types.OpRename, types.OpRead, types.OpOpen, types.OpClose}
	filter := NewEventFilterComponent(ops)

	for _, op := range ops {
		if !filter.ShouldWatch(op) {
			t.Fatalf("ShouldWatch(%v) = false, want true", op)
		}
	}
	if filter.ShouldWatch(types.OpChmod) {
		t.Fatal("ShouldWatch(OpChmod) = true, want false")
	}
}
