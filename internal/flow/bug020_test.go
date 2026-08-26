package flow

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"searchengine/internal/flowstate"
)

func TestShutdownReturnsWhenWorkerDeadlineExpires(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	var workers sync.WaitGroup
	workers.Add(1)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- coordinator.WaitForWorkers(ctx, &workers) }()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) || state.Status("workers") != "timeout" {
			t.Fatalf("shutdown returned the wrong deadline result or worker status: err=%v status=%s", err, state.Status("workers"))
		}
	case <-time.After(80 * time.Millisecond):
		t.Fatalf("shutdown ignored its deadline while a worker remained blocked")
	}
}
