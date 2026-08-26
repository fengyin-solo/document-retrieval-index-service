package flow

import (
	"context"
	"errors"
	"testing"
	"time"

	"searchengine/internal/flowstate"
)

func TestCorrectionCacheStopsCanceledLoadsAndExpiresEntries(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	_, canceledErr := coordinator.CorrectCached(canceled, "serch", time.Minute, time.Now, func(ctx context.Context, _ string) (string, error) {
		return "", ctx.Err()
	})
	clock := time.Unix(100, 0)
	loads := 0
	load := func(context.Context, string) (string, error) {
		loads++
		if loads == 1 {
			return "search-v1", nil
		}
		return "search-v2", nil
	}
	first, _ := coordinator.CorrectCached(context.Background(), "searchh", time.Minute, func() time.Time { return clock }, load)
	clock = clock.Add(2 * time.Minute)
	second, _ := coordinator.CorrectCached(context.Background(), "searchh", time.Minute, func() time.Time { return clock }, load)
	if !errors.Is(canceledErr, context.Canceled) || first != "search-v1" || second != "search-v2" || loads != 2 {
		t.Fatalf("correction work ignored cancellation or served an expired cached value: canceled=%v first=%q second=%q loads=%d", canceledErr, first, second, loads)
	}
}
