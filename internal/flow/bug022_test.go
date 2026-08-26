package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestAnalyzerGenerationsAreIndependentByName(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	english := coordinator.ReloadAnalyzer("english", 2, []string{"the"})
	code := coordinator.ReloadAnalyzer("code", 1, []string{"func"})
	englishState := state.Generation("analyzer:english")
	codeState := state.Generation("analyzer:code")
	if !english || !code || englishState.Generation != 2 || codeState.Generation != 1 || !reflect.DeepEqual(englishState.Values, []string{"the"}) || !reflect.DeepEqual(codeState.Values, []string{"func"}) {
		t.Fatalf("one analyzer generation blocked or replaced another analyzer: english=%+v code=%+v accepted=%v/%v", englishState, codeState, english, code)
	}
}
