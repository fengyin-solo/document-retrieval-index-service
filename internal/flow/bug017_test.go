package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestDeferredQueryLogOwnsItsTags(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	record := flowstate.QueryRecord{Query: "go", Tenant: "team-a", Tags: []string{"public"}}
	var job func()
	coordinator.LogQuery(record, func(run func()) { job = run })
	record.Tags[0] = "private"
	job()
	firstRead := state.Logs()
	firstRead[0].Tags[0] = "reader-change"
	secondRead := state.Logs()
	if len(secondRead) != 1 || !reflect.DeepEqual(secondRead[0].Tags, []string{"public"}) {
		t.Fatalf("deferred query log tags changed after request reuse or log reading: logs=%v", secondRead)
	}
}
