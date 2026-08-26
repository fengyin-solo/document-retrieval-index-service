package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestReindexReplacesTermsWithoutGrowingDocumentCount(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	coordinator.ReplaceDocument("main", "doc-1", []string{"old"})
	coordinator.ReplaceDocument("main", "doc-1", []string{"new"})
	terms := state.Snapshot("document:main:doc-1")
	if !reflect.DeepEqual(terms, []string{"new"}) || state.Counter("documents:main") != 1 {
		t.Fatalf("reindex appended stale terms or counted one document twice: terms=%v count=%d", terms, state.Counter("documents:main"))
	}
}
