package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestResultCacheReplacesPerFullQuery(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	coordinator.CacheResults("go", []string{"old-hit"})
	latest := coordinator.CacheResults("go", []string{"new-hit"})
	graph := coordinator.CacheResults("graph", []string{"graph-hit"})
	goCached := state.Snapshot("results:go")
	graphCached := state.Snapshot("results:graph")
	if !reflect.DeepEqual(latest, []string{"new-hit"}) || !reflect.DeepEqual(graph, []string{"graph-hit"}) || !reflect.DeepEqual(goCached, []string{"new-hit"}) || !reflect.DeepEqual(graphCached, []string{"graph-hit"}) {
		t.Fatalf("result cache appended stale hits or collided on a query prefix: latest=%v graph=%v goCached=%v graphCached=%v", latest, graph, goCached, graphCached)
	}
}
