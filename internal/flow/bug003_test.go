package flow

import (
	"context"
	"errors"
	"testing"

	"searchengine/internal/flowstate"
)

func TestTemporaryFetchRetriesBeforeReady(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	calls := 0
	value, err := coordinator.RunRetry(context.Background(), "synonyms", 3, func(_ context.Context, attempt int) (string, error) {
		calls++
		if attempt == 1 {
			return "", &TemporaryError{Err: errors.New("backend warming")}
		}
		return "loaded", nil
	})
	if err != nil || value != "loaded" || calls != 2 || state.Status("retry:synonyms") != "ready" {
		t.Fatalf("temporary fetch was not retried into a ready result: value=%q err=%v calls=%d status=%s", value, err, calls, state.Status("retry:synonyms"))
	}
}
