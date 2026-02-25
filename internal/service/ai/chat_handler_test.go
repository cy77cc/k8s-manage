package ai

import "testing"

func TestToolEventTrackerSummary(t *testing.T) {
	tracker := newToolEventTracker()
	tracker.noteCall("c1", "os.get_cpu_mem")
	tracker.noteCall("c2", "os.get_cpu_mem")
	tracker.noteCall("c3", "k8s.get_events")
	tracker.noteResult("c1", "os.get_cpu_mem")

	summary := tracker.summary()
	if summary.Calls != 3 {
		t.Fatalf("expected 3 calls, got %d", summary.Calls)
	}
	if summary.Results != 1 {
		t.Fatalf("expected 1 result, got %d", summary.Results)
	}
	if len(summary.Missing) != 2 {
		t.Fatalf("expected 2 missing results, got %d", len(summary.Missing))
	}
	if len(summary.MissingCallIDs) != 2 {
		t.Fatalf("expected 2 missing call ids, got %d", len(summary.MissingCallIDs))
	}
}

func TestResolveStreamState(t *testing.T) {
	ok := resolveStreamState(nil, toolSummary{})
	if ok != "ok" {
		t.Fatalf("expected ok state, got %s", ok)
	}

	partial := resolveStreamState(nil, toolSummary{MissingCallIDs: []string{"c1"}})
	if partial != "partial" {
		t.Fatalf("expected partial state, got %s", partial)
	}

	failed := resolveStreamState(&streamErrorPayload{
		Code:        "stream_interrupted",
		Message:     "broken",
		Recoverable: true,
	}, toolSummary{})
	if failed != "failed" {
		t.Fatalf("expected failed state, got %s", failed)
	}
}
