package flow

import (
	"reflect"
	"testing"

	"searchengine/internal/flowstate"
)

func TestBoostRulesReplacePerIndexSnapshot(t *testing.T) {
	state := flowstate.New()
	coordinator := New(state)
	coordinator.BoostSnapshot("manuals", []string{"title:2"})
	coordinator.BoostSnapshot("articles", []string{"body:3"})
	latest := coordinator.BoostSnapshot("manuals", []string{"title:4"})
	manuals := state.Snapshot("boost:manuals")
	articles := state.Snapshot("boost:articles")
	if !reflect.DeepEqual(latest, []string{"title:4"}) || !reflect.DeepEqual(manuals, []string{"title:4"}) || !reflect.DeepEqual(articles, []string{"body:3"}) {
		t.Fatalf("boost rule snapshots were appended or shared across indexes: latest=%v manuals=%v articles=%v", latest, manuals, articles)
	}
}
