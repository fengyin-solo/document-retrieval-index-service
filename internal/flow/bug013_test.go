package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestRepeatedDeleteDoesNotReapplyIndexSideEffects(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	coordinator.ReplaceDocument("main", "doc-1", []string{"term"})
	first := coordinator.DeleteDocument("main", "doc-1")
	second := coordinator.DeleteDocument("main", "doc-1")
	if !reflect.DeepEqual(first, []string{"term"}) || second != nil || state.HasSnapshot("document:main:doc-1") || state.Counter("documents:main") != 0 {
		t.Fatalf("repeated delete kept the document snapshot or decremented the index twice: first=%v second=%v count=%d", first, second, state.Counter("documents:main"))
	}
}
