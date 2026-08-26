package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestOlderRefreshCannotReplaceNewGeneration(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	fresh := []string{"fresh-term"}
	stale := []string{"stale-term"}
	first := coordinator.PublishGeneration("knowledge", 2, fresh)
	second := coordinator.PublishGeneration("knowledge", 1, stale)
	stale[0] = "mutated-stale"
	current := state.Generation("refresh:knowledge")
	if !first || second || current.Generation != 2 || !reflect.DeepEqual(current.Values, []string{"fresh-term"}) {
		t.Fatalf("older refresh replaced or polluted the current generation: first=%v second=%v generation=%d values=%v", first, second, current.Generation, current.Values)
	}
}
