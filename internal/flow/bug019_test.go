package flow

import (
	"testing"
	"time"

	"searchengine/internal/flowstate"
)

func TestRateWindowsAndClientsStayIndependent(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	start := time.Unix(100, 0)
	firstA := coordinator.AllowRequest("client-a", start, time.Minute, 1)
	secondA := coordinator.AllowRequest("client-a", start.Add(time.Second), time.Minute, 1)
	firstB := coordinator.AllowRequest("client-b", start.Add(time.Second), time.Minute, 1)
	nextA := coordinator.AllowRequest("client-a", start.Add(2*time.Minute), time.Minute, 1)
	if !firstA || secondA || !firstB || !nextA {
		t.Fatalf("rate limiter mixed clients or failed to open a new window: firstA=%v secondA=%v firstB=%v nextA=%v", firstA, secondA, firstB, nextA)
	}
}
