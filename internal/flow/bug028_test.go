package flow

import (
	"errors"
	"io"
	"strings"
	"testing"

	"searchengine/internal/flowstate"
)

type closeErrorReader struct {
	io.Reader
	closed *bool
}

func (r *closeErrorReader) Close() error {
	*r.closed = true
	return errors.New("reader close failed")
}

func TestTokenizerReportsReaderCloseFailure(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	closed := false
	err := coordinator.TokenizeReaders([]string{"doc"}, func(string) (io.ReadCloser, error) {
		return &closeErrorReader{Reader: strings.NewReader("text"), closed: &closed}, nil
	}, func(io.Reader) error { return nil })
	if err == nil || !closed || state.Status("tokenize") != "failed" {
		t.Fatalf("tokenizer hid the reader close failure or published success: err=%v closed=%v status=%s", err, closed, state.Status("tokenize"))
	}
}
