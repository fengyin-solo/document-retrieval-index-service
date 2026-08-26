package flow

import (
	"errors"
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestFailedMultiIndexRemovalKeepsPriorCheckpoint(t *testing.T) {
	state := flowstate.New()
	state.PutSnapshot("removed:doc-1", []string{"previous"})
	coordinator := New(state)
	err := coordinator.RemoveFromIndexes("doc-1", []string{"main", "archive"}, func(indexID, _ string) error {
		if indexID == "archive" {
			return errors.New("archive unavailable")
		}
		return nil
	})
	if err == nil || !reflect.DeepEqual(state.Snapshot("removed:doc-1"), []string{"previous"}) || state.Status("remove:doc-1") != "failed" {
		t.Fatalf("failed multi-index removal published a partial checkpoint or unfinished status: err=%v checkpoint=%v status=%s", err, state.Snapshot("removed:doc-1"), state.Status("remove:doc-1"))
	}
}
