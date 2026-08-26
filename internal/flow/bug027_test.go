package flow

import (
	"testing"

	"searchengine/internal/flowstate"
)

func TestRecoveredSearchDoesNotPublishPartialResults(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	results, err := coordinator.RecoverSearch(func(values *[]string) {
		*values = append(*values, "partial-hit")
		panic("scorer failed")
	})
	if err == nil || results != nil || len(state.Snapshot("search")) != 0 || state.Status("search") != "failed" {
		t.Fatalf("recovered search returned or cached partial hits: results=%v err=%v cached=%v status=%s", results, err, state.Snapshot("search"), state.Status("search"))
	}
}
