package component

import (
	"testing"

	"binmonitor/internal/types"
)

func TestEventFilterComponentShouldWatchConfiguredEvents(t *testing.T) {
	filter, err := NewEventFilterComponent([]string{"create", "modify", "delete", "read"})
	if err != nil {
		t.Fatalf("NewEventFilterComponent() error = %v", err)
	}

	for _, op := range []types.FileOp{types.OpCreate, types.OpWrite, types.OpRemove, types.OpRename, types.OpRead} {
		if !filter.ShouldWatch(op) {
			t.Fatalf("ShouldWatch(%v) = false, want true", op)
		}
	}
	if filter.ShouldWatch(types.OpChmod) {
		t.Fatal("ShouldWatch(OpChmod) = true, want false")
	}
}

func TestEventFilterComponentRejectsUnknownEvent(t *testing.T) {
	if _, err := NewEventFilterComponent([]string{"unknown"}); err == nil {
		t.Fatal("NewEventFilterComponent() error = nil, want error")
	}
}
