package flow

import (
	"context"
	"reflect"
	"testing"
	"time"

	"searchengine/internal/flowstate"
)

func TestFanoutClosesAfterAllShardResults(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	done := make(chan []string, 1)
	go func() {
		values, _ := coordinator.Fanout(context.Background(), []string{"b", "a"}, func(_ context.Context, shard string) (string, error) {
			return shard + "-result", nil
		})
		done <- values
	}()
	select {
	case values := <-done:
		if !reflect.DeepEqual(values, []string{"a-result", "b-result"}) || state.Status("fanout") != "ready" {
			t.Fatalf("fanout completed without all sorted shard results and ready status: values=%v status=%s", values, state.Status("fanout"))
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("fanout waited forever after every shard had returned")
	}
}
