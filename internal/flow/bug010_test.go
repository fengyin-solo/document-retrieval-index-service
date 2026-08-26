package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestCapturedRequestFiltersStayWithOriginalTenant(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	meta := &RequestMeta{Tenant: "team-a", Filters: []string{"public"}}
	var consumed RequestMeta
	coordinator.CaptureRequest(meta, func(snapshot RequestMeta) { consumed = snapshot })
	meta.Tenant = "team-b"
	meta.Filters[0] = "private"
	stored := state.Snapshot("request:team-a")
	if consumed.Tenant != "team-a" || !reflect.DeepEqual(consumed.Filters, []string{"public"}) || !reflect.DeepEqual(stored, []string{"public"}) {
		t.Fatalf("reused request metadata changed the prior tenant filters: consumed=%+v stored=%v", consumed, stored)
	}
}
