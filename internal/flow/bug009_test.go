package flow

import (
	"context"
	"errors"
	"testing"

	"searchengine/internal/flowstate"
)

func TestAutocompleteStopsAfterRequestCancellation(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	values, err := coordinator.Autocomplete(ctx, []string{"go", "golang"}, func(_ context.Context, input string) (string, error) {
		calls++
		if calls == 1 {
			cancel()
		}
		return input + "-suggestion", nil
	})
	if !errors.Is(err, context.Canceled) || values != nil || calls != 1 || len(state.Snapshot("autocomplete")) != 0 || state.Status("autocomplete") != "canceled" {
		t.Fatalf("autocomplete continued and saved results after cancellation: values=%v err=%v calls=%d saved=%v status=%s", values, err, calls, state.Snapshot("autocomplete"), state.Status("autocomplete"))
	}
}
