package flow

import (
	"testing"

	"searchengine/internal/flowstate"
)

func TestPanickingBuildNeverPublishesPartialIndex(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	err := coordinator.BuildAtomic("manuals", func(values *[]string) {
		*values = append(*values, "partial-term")
		panic("tokenizer failed")
	})
	if err == nil || len(state.Snapshot("build:manuals")) != 0 || state.Status("build:manuals") != "failed" {
		t.Fatalf("panicking build published a partial index or kept a building status: err=%v snapshot=%v status=%s", err, state.Snapshot("build:manuals"), state.Status("build:manuals"))
	}
}
