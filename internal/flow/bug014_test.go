package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestSynonymSnapshotIsolatedFromEditors(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	input := []string{"car", "vehicle"}
	returned := coordinator.StoreSynonyms("auto", input)
	input[0] = "caller-change"
	returned[1] = "reader-change"
	if stored := state.Snapshot("synonyms:auto"); !reflect.DeepEqual(stored, []string{"car", "vehicle"}) {
		t.Fatalf("synonym editors changed the active synonym snapshot: stored=%v", stored)
	}
}
