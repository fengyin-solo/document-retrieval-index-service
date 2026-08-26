package flow

import (
	"context"
	"errors"
	"testing"
	"time"

	"searchengine/internal/flowstate"
)

func TestCanceledLookupStopsAndMarksFailure(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	_, err := coordinator.DeadlineLookup(ctx, "manual", func(callCtx context.Context, _ string) (string, error) {
		select {
		case <-callCtx.Done():
			return "", callCtx.Err()
		case <-time.After(80 * time.Millisecond):
			return "", errors.New("lookup kept running")
		}
	})
	if !errors.Is(err, context.Canceled) || time.Since(started) > 40*time.Millisecond || state.Status("lookup:manual") != "failed" {
		t.Fatalf("canceled lookup kept the backend alive or left a ready status: err=%v elapsed=%s status=%s", err, time.Since(started), state.Status("lookup:manual"))
	}
}
