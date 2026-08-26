package flow

import (
	"errors"
	"io"
	"strings"
	"testing"

	"searchengine/internal/flowstate"
)

type countedReader struct {
	io.Reader
	open *int
}

func (r *countedReader) Close() error {
	*r.open--
	return nil
}

func TestBatchReadersCloseBeforeNextOpen(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	openCount := 0
	count, err := coordinator.ReadBatch("docs", []string{"a", "b", "c"}, func(_ string) (io.ReadCloser, error) {
		if openCount >= 1 {
			return nil, errors.New("reader limit reached")
		}
		openCount++
		return &countedReader{Reader: strings.NewReader("body"), open: &openCount}, nil
	})
	if err != nil || count != 3 || openCount != 0 || state.Status("batch:docs") != "ready" {
		t.Fatalf("batch held readers across items or left the batch unfinished: count=%d open=%d err=%v status=%s", count, openCount, err, state.Status("batch:docs"))
	}
}
