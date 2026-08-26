package flow

import (
	"errors"
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestFailedRebuildKeepsPreviousSearchSnapshot(t *testing.T) {
	state := flowstate.New()
	state.PutSnapshot("rebuild:main", []string{"stable-term"})
	coordinator := New(state)
	err := coordinator.RebuildTransactional("main", func() ([]string, error) {
		return []string{"partial-term"}, errors.New("source unavailable")
	})
	if err == nil || !reflect.DeepEqual(state.Snapshot("rebuild:main"), []string{"stable-term"}) || state.Status("rebuild:main") != "failed" {
		t.Fatalf("failed rebuild erased the previous snapshot or kept a running state: err=%v snapshot=%v status=%s", err, state.Snapshot("rebuild:main"), state.Status("rebuild:main"))
	}
}
