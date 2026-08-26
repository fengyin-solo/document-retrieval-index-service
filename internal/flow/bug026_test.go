package flow

import (
	"context"
	"errors"
	"testing"

	"searchengine/internal/flowstate"
)

func TestIdempotencyKeyCommitsOnlyAfterSuccessfulWrite(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	calls := 0
	firstErr := coordinator.RetryWrite(context.Background(), "doc-a", 1, func(context.Context) error {
		calls++
		return &TemporaryError{Err: errors.New("temporary")}
	})
	secondErr := coordinator.RetryWrite(context.Background(), "doc-a", 1, func(context.Context) error {
		calls++
		return nil
	})
	thirdErr := coordinator.RetryWrite(context.Background(), "doc-b", 1, func(context.Context) error {
		calls++
		return nil
	})
	if firstErr == nil || secondErr != nil || thirdErr != nil || calls != 3 || !state.Seen("doc-a") || !state.Seen("doc-b") {
		t.Fatalf("failed write consumed its idempotency key or another key: first=%v second=%v third=%v calls=%d", firstErr, secondErr, thirdErr, calls)
	}
}
