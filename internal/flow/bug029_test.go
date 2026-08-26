package flow

import (
	"context"
	"errors"
	"testing"
	"time"

	"searchengine/internal/flowstate"
)

func TestQueueReturnsFirstConsumerErrorWithoutBlockingProducer(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	done := make(chan error, 1)
	go func() {
		done <- coordinator.QueueIndex(context.Background(), []string{"doc-1", "doc-2"}, func(context.Context, string) error {
			return errors.New("consumer stopped")
		})
	}()
	select {
	case err := <-done:
		if err == nil || state.Status("index-queue") != "failed" {
			t.Fatalf("index queue lost the consumer error or final failure state: err=%v status=%s", err, state.Status("index-queue"))
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("index queue producer blocked after the consumer returned an error")
	}
}
