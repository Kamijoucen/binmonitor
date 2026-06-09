package logic

import (
	"testing"

	"binmonitor/internal/types"
)

func TestResolveEventOps(t *testing.T) {
	tests := []struct {
		name    string
		want    []types.FileOp
		wantErr bool
	}{
		{"create", []types.FileOp{types.OpCreate}, false},
		{"write", []types.FileOp{types.OpWrite}, false},
		{"modify", []types.FileOp{types.OpWrite}, false},
		{"MODIFY", []types.FileOp{types.OpWrite}, false},
		{"remove", []types.FileOp{types.OpRemove}, false},
		{"delete", []types.FileOp{types.OpRemove, types.OpRename}, false},
		{"rename", []types.FileOp{types.OpRename}, false},
		{"read", []types.FileOp{types.OpRead}, false},
		{"open", []types.FileOp{types.OpOpen}, false},
		{"process_open", []types.FileOp{types.OpOpen}, false},
		{"close", []types.FileOp{types.OpClose}, false},
		{"process_close", []types.FileOp{types.OpClose}, false},
		{"unknown", nil, true},
		{"", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveEventOps(tt.name)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveEventOps(%q) expected error", tt.name)
				}
				return
			}
			if err != nil {
				t.Errorf("ResolveEventOps(%q) unexpected error: %v", tt.name, err)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("ResolveEventOps(%q) = %v, want %v", tt.name, got, tt.want)
				return
			}
			for i, op := range got {
				if op != tt.want[i] {
					t.Errorf("ResolveEventOps(%q)[%d] = %v, want %v", tt.name, i, op, tt.want[i])
				}
			}
		})
	}
}

func TestResolveAllEventOps(t *testing.T) {
	ops, err := ResolveAllEventOps([]string{"create", "modify", "delete"})
	if err != nil {
		t.Fatalf("ResolveAllEventOps() error = %v", err)
	}
	// create→OpCreate, modify→OpWrite, delete→OpRemove+OpRename = 4 ops, 去重后 4
	if len(ops) != 4 {
		t.Errorf("ResolveAllEventOps len = %d, want 4", len(ops))
	}
	seen := make(map[types.FileOp]bool)
	for _, op := range ops {
		seen[op] = true
	}
	if !seen[types.OpCreate] || !seen[types.OpWrite] || !seen[types.OpRemove] || !seen[types.OpRename] {
		t.Errorf("ResolveAllEventOps missing expected ops: %v", ops)
	}
}

func TestResolveAllEventOpsDuplicate(t *testing.T) {
	ops, err := ResolveAllEventOps([]string{"write", "modify"})
	if err != nil {
		t.Fatalf("ResolveAllEventOps() error = %v", err)
	}
	// write→OpWrite, modify→OpWrite, 去重后只有 OpWrite
	if len(ops) != 1 {
		t.Errorf("ResolveAllEventOps len = %d, want 1 (dedup)", len(ops))
	}
	if ops[0] != types.OpWrite {
		t.Errorf("ResolveAllEventOps[0] = %v, want OpWrite", ops[0])
	}
}

func TestResolveAllEventOpsRejectsUnknown(t *testing.T) {
	if _, err := ResolveAllEventOps([]string{"unknown"}); err == nil {
		t.Fatal("ResolveAllEventOps() error = nil, want error")
	}
}
