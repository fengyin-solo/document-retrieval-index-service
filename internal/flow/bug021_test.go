package flow

import (
	"context"
	"errors"
	"testing"

	"searchengine/internal/flowstate"
)

func TestShardQueryStopsBeforeNextShardAfterCancel(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	values, err := coordinator.QueryShards(ctx, []string{"primary", "archive"}, func(_ context.Context, shard string) (string, error) {
		calls++
		if shard == "primary" {
			cancel()
		}
		return shard + "-result", nil
	})
	if !errors.Is(err, context.Canceled) || values != nil || calls != 1 || len(state.Snapshot("shard-query")) != 0 || state.Status("shard-query") != "canceled" {
		t.Fatalf("canceled shard query continued into another shard or saved results: values=%v err=%v calls=%d saved=%v status=%s", values, err, calls, state.Snapshot("shard-query"), state.Status("shard-query"))
	}
}
