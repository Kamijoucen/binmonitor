package component

import (
	"testing"

	"binmonitor/internal/types"
)

func TestDiffProcessFDSnapshotReportsOpenCloseAndReuse(t *testing.T) {
	previous := processFDSnapshot{
		3: "/tmp/old.txt",
		4: "/tmp/keep.txt",
		5: "/tmp/reused-old.txt",
	}
	next := processFDSnapshot{
		4: "/tmp/keep.txt",
		5: "/tmp/reused-new.txt",
		6: "/tmp/new.txt",
	}

	events := diffProcessFDSnapshot(123, "worker", previous, next)
	if len(events) != 4 {
		t.Fatalf("event count = %d, want 4: %#v", len(events), events)
	}
	want := []types.FileEvent{
		{Path: "/tmp/old.txt", Op: types.OpClose, PID: 123, FD: 3, ProcessName: "worker"},
		{Path: "/tmp/reused-old.txt", Op: types.OpClose, PID: 123, FD: 5, ProcessName: "worker"},
		{Path: "/tmp/reused-new.txt", Op: types.OpOpen, PID: 123, FD: 5, ProcessName: "worker"},
		{Path: "/tmp/new.txt", Op: types.OpOpen, PID: 123, FD: 6, ProcessName: "worker"},
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events[%d] = %#v, want %#v", i, events[i], want[i])
		}
	}
}

func TestIsProcessFileTarget(t *testing.T) {
	for _, target := range []string{"/tmp/file.txt", "/tmp/file.txt (deleted)"} {
		if !isProcessFileTarget(target) {
			t.Fatalf("isProcessFileTarget(%q) = false, want true", target)
		}
	}
	for _, target := range []string{"", "socket:[1]", "pipe:[1]", "anon_inode:[eventfd]", "memfd:jit-cache"} {
		if isProcessFileTarget(target) {
			t.Fatalf("isProcessFileTarget(%q) = true, want false", target)
		}
	}
}
