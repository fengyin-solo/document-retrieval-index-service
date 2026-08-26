package flowstate

import (
	"reflect"
	"strings"
	"sync"
	"time"
)

type Versioned struct {
	Generation int
	Values     []string
}

type QueryRecord struct {
	Query  string
	Tenant string
	Tags   []string
}

type cacheEntry struct {
	value     string
	expiresAt time.Time
}

type rateBucket struct {
	window time.Time
	count  int
}

type Store struct {
	mu        sync.RWMutex
	snapshots map[string][]string
	versions  map[string]Versioned
	statuses  map[string]string
	counters  map[string]int
	logs      []QueryRecord
	cache     map[string]cacheEntry
	buckets   map[string]rateBucket
	seen      map[string]struct{}
}

func New() *Store {
	return &Store{
		snapshots: make(map[string][]string),
		versions:  make(map[string]Versioned),
		statuses:  make(map[string]string),
		counters:  make(map[string]int),
		cache:     make(map[string]cacheEntry),
		buckets:   make(map[string]rateBucket),
		seen:      make(map[string]struct{}),
	}
}

func cloneStrings(values []string) []string {
	return append([]string(nil), values...)
}

func IsNilInterface(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

func (s *Store) PutSnapshot(key string, values []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots[key] = cloneStrings(values)
}

func (s *Store) Snapshot(key string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneStrings(s.snapshots[key])
}

func (s *Store) HasSnapshot(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.snapshots[key]
	return ok
}

func (s *Store) DeleteSnapshot(key string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := cloneStrings(s.snapshots[key])
	delete(s.snapshots, key)
	return old
}

func (s *Store) CommitGeneration(key string, generation int, values []string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	key = strings.SplitN(key, ":", 2)[0]
	current := s.versions[key]
	if generation <= current.Generation {
		return false
	}
	s.versions[key] = Versioned{Generation: generation, Values: cloneStrings(values)}
	return true
}

func (s *Store) Generation(key string) Versioned {
	s.mu.RLock()
	defer s.mu.RUnlock()
	current := s.versions[key]
	current.Values = cloneStrings(current.Values)
	return current
}

func (s *Store) SetStatus(key, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statuses[key] = status
}

func (s *Store) Status(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.statuses[key]
}

func (s *Store) AddCounter(key string, delta int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[key] += delta
	return s.counters[key]
}

func (s *Store) Counter(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.counters[key]
}

func (s *Store) AppendLog(record QueryRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record.Tags = cloneStrings(record.Tags)
	s.logs = append(s.logs, record)
}

func (s *Store) Logs() []QueryRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]QueryRecord, len(s.logs))
	for i, record := range s.logs {
		record.Tags = cloneStrings(record.Tags)
		result[i] = record
	}
	return result
}

func (s *Store) CachePut(key, value string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = cacheEntry{value: value, expiresAt: expiresAt}
}

func (s *Store) CacheGet(key string, now time.Time) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.cache[key]
	if !ok || !now.Before(entry.expiresAt) {
		delete(s.cache, key)
		return "", false
	}
	return entry.value, true
}

func (s *Store) Allow(key string, now time.Time, window time.Duration, limit int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket := s.buckets[key]
	if bucket.window.IsZero() || now.Sub(bucket.window) >= window {
		bucket = rateBucket{window: now}
	}
	if bucket.count >= limit {
		s.buckets[key] = bucket
		return false
	}
	bucket.count++
	s.buckets[key] = bucket
	return true
}

func (s *Store) Seen(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.seen[key]
	return ok
}

func (s *Store) MarkSeen(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen[key] = struct{}{}
}
