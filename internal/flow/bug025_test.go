package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestFacetCacheSurvivesScratchSliceReuse(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	counts := coordinator.FacetSnapshot(map[string]string{"doc-1": "guides", "doc-2": "news"})
	cached := state.Snapshot("facets")
	if !reflect.DeepEqual(counts, map[string]int{"guides": 1, "news": 1}) || !reflect.DeepEqual(cached, []string{"guides", "news"}) {
		t.Fatalf("facet category cache was cleared when its scratch slice was reused: counts=%v cached=%v", counts, cached)
	}
}
