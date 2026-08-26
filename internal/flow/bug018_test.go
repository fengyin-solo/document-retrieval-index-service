package flow

import (
	"context"
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestSuggestionMergeWaitsForEveryProvider(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	providers := []func(context.Context) ([]string, error){
		func(context.Context) ([]string, error) { return []string{"alpha", "common"}, nil },
		func(context.Context) ([]string, error) { return []string{"beta", "common"}, nil },
	}
	merged, err := coordinator.MergeSuggestions(context.Background(), providers)
	want := []string{"alpha", "beta", "common"}
	if err != nil || !reflect.DeepEqual(merged, want) || !reflect.DeepEqual(state.Snapshot("suggestions"), want) || state.Status("suggestions") != "ready" {
		t.Fatalf("suggestion merge returned a partial provider result or unfinished status: merged=%v saved=%v err=%v status=%s", merged, state.Snapshot("suggestions"), err, state.Status("suggestions"))
	}
}
