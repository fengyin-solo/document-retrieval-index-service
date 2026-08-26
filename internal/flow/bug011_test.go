package flow

import (
	"testing"

	"searchengine/internal/flowstate"
)

func TestHotTermsKeepIndependentCounters(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	coordinator.RecordHot("alpha")
	coordinator.RecordHot("beta")
	if state.Counter("hot:alpha") != 1 || state.Counter("hot:beta") != 1 {
		t.Fatalf("different hot terms shared the first term counter: alpha=%d beta=%d", state.Counter("hot:alpha"), state.Counter("hot:beta"))
	}
}
