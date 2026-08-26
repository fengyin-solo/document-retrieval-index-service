package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestExportSnapshotSurvivesCallerReuse(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	batch := []string{"alpha", "beta"}
	exported := coordinator.StableExport("daily", batch)
	batch[0] = "second-request"
	exported[1] = "consumer-change"
	stored := state.Snapshot("daily")
	if !reflect.DeepEqual(stored, []string{"alpha", "beta"}) {
		t.Fatalf("first export snapshot was changed by later buffer reuse: got=%v", stored)
	}
}
