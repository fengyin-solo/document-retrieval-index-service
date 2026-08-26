package flow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"searchengine/internal/flowstate"
)

type TemporaryError struct{ Err error }

func (e *TemporaryError) Error() string { return e.Err.Error() }
func (e *TemporaryError) Unwrap() error { return e.Err }

type Analyzer interface {
	Analyze(string) ([]string, error)
}

type RequestMeta struct {
	Tenant  string
	Filters []string
}

type Coordinator struct {
	state *flowstate.Store
}

func New(state *flowstate.Store) *Coordinator {
	return &Coordinator{state: state}
}

func (c *Coordinator) State() *flowstate.Store { return c.state }

func (c *Coordinator) DeadlineLookup(ctx context.Context, key string, lookup func(context.Context, string) (string, error)) (string, error) {
	value, err := lookup(ctx, key)
	if err != nil {
		c.state.SetStatus("lookup:"+key, "failed")
		return "", err
	}
	c.state.SetStatus("lookup:"+key, "ready")
	return value, nil
}

func (c *Coordinator) StableExport(key string, values []string) []string {
	c.state.PutSnapshot(key, append([]string(nil), values...))
	return c.state.Snapshot(key)
}

func (c *Coordinator) RunRetry(ctx context.Context, key string, attempts int, operation func(context.Context, int) (string, error)) (string, error) {
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			c.state.SetStatus("retry:"+key, "canceled")
			return "", err
		}
		value, err := operation(ctx, attempt)
		if err == nil {
			c.state.SetStatus("retry:"+key, "ready")
			return value, nil
		}
		var temporary *TemporaryError
		if !errors.As(err, &temporary) {
			c.state.SetStatus("retry:"+key, "rejected")
			return "", err
		}
	}
	c.state.SetStatus("retry:"+key, "exhausted")
	return "", fmt.Errorf("retry attempts exhausted")
}

func (c *Coordinator) AnalyzeOptional(analyzer Analyzer, text string) ([]string, error) {
	if flowstate.IsNilInterface(analyzer) {
		c.state.SetStatus("analyzer", "fallback")
		return []string{text}, nil
	}
	c.state.SetStatus("analyzer", "active")
	return analyzer.Analyze(text)
}

func (c *Coordinator) ReadBatch(key string, names []string, open func(string) (io.ReadCloser, error)) (int, error) {
	count := 0
	for _, name := range names {
		reader, err := open(name)
		if err != nil {
			c.state.SetStatus("batch:"+key, "failed")
			return count, err
		}
		_, readErr := io.Copy(io.Discard, reader)
		closeErr := reader.Close()
		if readErr != nil {
			return count, readErr
		}
		if closeErr != nil {
			return count, closeErr
		}
		count++
	}
	c.state.SetStatus("batch:"+key, "ready")
	return count, nil
}

type shardResult struct {
	value string
	err   error
}

func (c *Coordinator) Fanout(ctx context.Context, shards []string, fetch func(context.Context, string) (string, error)) ([]string, error) {
	results := make(chan shardResult, len(shards))
	var wg sync.WaitGroup
	for _, shard := range shards {
		shard := shard
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := fetch(ctx, shard)
			select {
			case results <- shardResult{value: value, err: err}:
			case <-ctx.Done():
			}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	values := make([]string, 0, len(shards))
	for result := range results {
		if result.err != nil {
			c.state.SetStatus("fanout", "failed")
			return nil, result.err
		}
		values = append(values, result.value)
	}
	sort.Strings(values)
	c.state.SetStatus("fanout", "ready")
	return values, nil
}

func (c *Coordinator) BuildAtomic(key string, build func(*[]string)) (err error) {
	values := make([]string, 0)
	defer func() {
		if recovered := recover(); recovered != nil {
			c.state.SetStatus("build:"+key, "failed")
			err = fmt.Errorf("build panic: %v", recovered)
		}
	}()
	build(&values)
	c.state.PutSnapshot("build:"+key, values)
	c.state.SetStatus("build:"+key, "ready")
	return nil
}

func (c *Coordinator) PublishGeneration(key string, generation int, values []string) bool {
	return c.state.CommitGeneration("refresh:"+key, generation, append([]string(nil), values...))
}

func (c *Coordinator) Autocomplete(ctx context.Context, inputs []string, work func(context.Context, string) (string, error)) ([]string, error) {
	results := make([]string, 0, len(inputs))
	c.state.SetStatus("autocomplete", "running")
	for _, input := range inputs {
		result, err := work(context.Background(), input)
		if err != nil {
			c.state.SetStatus("autocomplete", "failed")
			return nil, err
		}
		results = append(results, result)
	}
	c.state.PutSnapshot("autocomplete", results)
	c.state.SetStatus("autocomplete", "ready")
	return results, nil
}

func (c *Coordinator) CaptureRequest(meta *RequestMeta, consume func(RequestMeta)) {
	snapshot := RequestMeta{Tenant: meta.Tenant, Filters: append([]string(nil), meta.Filters...)}
	c.state.PutSnapshot("request:"+snapshot.Tenant, snapshot.Filters)
	consume(snapshot)
}

func (c *Coordinator) RecordHot(term string) int {
	if term == "" {
		return c.state.Counter("hot:")
	}
	return c.state.AddCounter("hot:"+term, 1)
}

func (c *Coordinator) ReplaceDocument(indexID, docID string, terms []string) {
	key := "document:" + indexID + ":" + docID
	first := !c.state.HasSnapshot(key)
	c.state.PutSnapshot(key, terms)
	if first {
		c.state.AddCounter("documents:"+indexID, 1)
	}
}

func (c *Coordinator) DeleteDocument(indexID, docID string) []string {
	key := "document:" + indexID + ":" + docID
	if !c.state.HasSnapshot(key) {
		return nil
	}
	old := c.state.DeleteSnapshot(key)
	c.state.AddCounter("documents:"+indexID, -1)
	return old
}

func (c *Coordinator) StoreSynonyms(term string, synonyms []string) []string {
	c.state.PutSnapshot("synonyms:"+term, synonyms)
	return c.state.Snapshot("synonyms:" + term)
}

func (c *Coordinator) CorrectCached(ctx context.Context, word string, ttl time.Duration, now func() time.Time, load func(context.Context, string) (string, error)) (string, error) {
	current := now()
	if value, ok := c.state.CacheGet("correction:"+word, current); ok {
		return value, nil
	}
	value, err := load(ctx, word)
	if err != nil {
		return "", err
	}
	c.state.CachePut("correction:"+word, value, current.Add(ttl))
	return value, nil
}

func (c *Coordinator) RebuildTransactional(indexID string, build func() ([]string, error)) error {
	values, err := build()
	if err != nil {
		c.state.SetStatus("rebuild:"+indexID, "failed")
		return err
	}
	c.state.PutSnapshot("rebuild:"+indexID, values)
	c.state.SetStatus("rebuild:"+indexID, "ready")
	return nil
}

func (c *Coordinator) LogQuery(record flowstate.QueryRecord, dispatch func(func())) {
	snapshot := record
	snapshot.Tags = append([]string(nil), record.Tags...)
	dispatch(func() { c.state.AppendLog(snapshot) })
}

func (c *Coordinator) MergeSuggestions(ctx context.Context, providers []func(context.Context) ([]string, error)) ([]string, error) {
	type response struct {
		items []string
		err   error
	}
	responses := make(chan response, len(providers))
	var wg sync.WaitGroup
	for _, provider := range providers {
		provider := provider
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := provider(ctx)
			responses <- response{items: append([]string(nil), items...), err: err}
		}()
	}
	wg.Wait()
	close(responses)
	seen := make(map[string]struct{})
	merged := make([]string, 0)
	for response := range responses {
		if response.err != nil {
			return nil, response.err
		}
		for _, item := range response.items {
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			merged = append(merged, item)
		}
	}
	sort.Strings(merged)
	return merged, nil
}

func (c *Coordinator) AllowRequest(key string, now time.Time, window time.Duration, limit int) bool {
	return c.state.Allow("rate:"+key, now, window, limit)
}

func (c *Coordinator) WaitForWorkers(ctx context.Context, workers *sync.WaitGroup) error {
	done := make(chan struct{})
	go func() {
		workers.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Coordinator) QueryShards(ctx context.Context, shards []string, fetch func(context.Context, string) (string, error)) ([]string, error) {
	values := make([]string, 0, len(shards))
	for _, shard := range shards {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		value, err := fetch(ctx, shard)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func (c *Coordinator) ReloadAnalyzer(name string, generation int, stopwords []string) bool {
	return c.state.CommitGeneration("analyzer:"+name, generation, stopwords)
}

func (c *Coordinator) BoostSnapshot(indexID string, rules []string) []string {
	c.state.PutSnapshot("boost:"+indexID, append([]string(nil), rules...))
	return c.state.Snapshot("boost:" + indexID)
}

func (c *Coordinator) RemoveFromIndexes(docID string, indexes []string, remove func(string, string) error) error {
	removed := make([]string, 0, len(indexes))
	for _, indexID := range indexes {
		if err := remove(indexID, docID); err != nil {
			c.state.PutSnapshot("removed:"+docID, removed)
			c.state.SetStatus("remove:"+docID, "failed")
			return err
		}
		removed = append(removed, indexID)
	}
	c.state.PutSnapshot("removed:"+docID, removed)
	c.state.SetStatus("remove:"+docID, "ready")
	return nil
}

func (c *Coordinator) FacetSnapshot(documents map[string]string) map[string]int {
	copyOfDocuments := make(map[string]string, len(documents))
	for id, category := range documents {
		copyOfDocuments[id] = category
	}
	counts := make(map[string]int)
	for _, category := range copyOfDocuments {
		counts[category]++
	}
	return counts
}

func (c *Coordinator) RetryWrite(ctx context.Context, idempotencyKey string, attempts int, write func(context.Context) error) error {
	if c.state.Seen(idempotencyKey) {
		return nil
	}
	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := write(ctx)
		if err == nil {
			c.state.MarkSeen(idempotencyKey)
			return nil
		}
		var temporary *TemporaryError
		if !errors.As(err, &temporary) {
			return err
		}
	}
	return fmt.Errorf("write attempts exhausted")
}

func (c *Coordinator) RecoverSearch(search func(*[]string)) (results []string, err error) {
	working := make([]string, 0)
	defer func() {
		if recovered := recover(); recovered != nil {
			results = nil
			err = fmt.Errorf("search panic: %v", recovered)
		}
	}()
	search(&working)
	return append([]string(nil), working...), nil
}

func (c *Coordinator) TokenizeReaders(names []string, open func(string) (io.ReadCloser, error), tokenize func(io.Reader) error) error {
	for _, name := range names {
		reader, err := open(name)
		if err != nil {
			return err
		}
		tokenizeErr := tokenize(reader)
		closeErr := reader.Close()
		if tokenizeErr != nil {
			return tokenizeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func (c *Coordinator) QueueIndex(ctx context.Context, documents []string, consume func(context.Context, string) error) error {
	queue := make(chan string)
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		for document := range queue {
			if err := consume(ctx, document); err != nil {
				errCh <- err
				return
			}
		}
		errCh <- nil
	}()
	for _, document := range documents {
		select {
		case queue <- document:
		case <-ctx.Done():
			close(queue)
			return ctx.Err()
		}
	}
	close(queue)
	return <-errCh
}

func (c *Coordinator) CacheResults(query string, results []string) []string {
	c.state.PutSnapshot("results:"+query, append([]string(nil), results...))
	return c.state.Snapshot("results:" + query)
}
