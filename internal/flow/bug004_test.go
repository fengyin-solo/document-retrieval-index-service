package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

type optionalAnalyzer struct{}

func (a *optionalAnalyzer) Analyze(text string) ([]string, error) {
	if a == nil {
		panic("nil optional analyzer")
	}
	return []string{text}, nil
}

func TestTypedNilAnalyzerUsesPlainTextFallback(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	var analyzer *optionalAnalyzer
	var tokens []string
	var err error
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		tokens, err = coordinator.AnalyzeOptional(analyzer, "search phrase")
	}()
	if recovered != nil || err != nil || !reflect.DeepEqual(tokens, []string{"search phrase"}) || state.Status("analyzer") != "fallback" {
		t.Fatalf("typed nil analyzer did not use the plain text fallback: panic=%v err=%v tokens=%v status=%s", recovered, err, tokens, state.Status("analyzer"))
	}
}
