package tag_paths

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/catgut"
)

func TestAddPaths(t1 *testing.T) {
	t := ui.T{T: t1}

	var es Tags

	es.AddPath(MakePathWithType(
		catgut.MakeFromString("area-home"),
		catgut.MakeFromString("project-reno"),
	))

	{
		i, ok := es.All.ContainsTag(catgut.MakeFromString("area"))

		if !ok {
			t.Errorf("expected some tag: %d, %t, %s", i, ok, es)
		}
	}

	es.AddPath(MakePathWithType(
		catgut.MakeFromString("area-career"),
		catgut.MakeFromString("project-recurse"),
	))

	{
		i, ok := es.All.ContainsTag(catgut.MakeFromString("area"))

		if !ok {
			t.Errorf("expected some tag: %d, %t, %s", i, ok, es.All)
		}
	}
}

func TestRealWorld(t1 *testing.T) {
	t := ui.T{T: t1}

	var es Tags

	es.AddTag(catgut.MakeFromString("pom-1"))
	es.AddTag(catgut.MakeFromString("req-comp-internet"))
	es.AddTag(catgut.MakeFromString("today-in_progress"))

	{
		e := catgut.MakeFromString("req-comp-internet")
		_, ok := es.All.ContainsTag(e)

		if !ok {
			t.Errorf("expected %s to be in %s", e, es)
		}
	}

	es.AddPath(MakePathWithType(
		catgut.MakeFromString("project-2022-recurse"),
		catgut.MakeFromString("project-24q2-talent_show"),
	))

	e := catgut.MakeFromString("req-comp-internet")
	_, ok := es.All.ContainsTag(e)

	if !ok {
		t.Errorf("expected %s to be in %s", e, es)
	}
}

func BenchmarkMatchFirstYes(b *testing.B) {
	var es Tags

	es.AddPath(MakePathWithType(
		catgut.MakeFromString("area-home"),
		catgut.MakeFromString("project-reno"),
	))

	es.AddPath(MakePathWithType(
		catgut.MakeFromString("area-career"),
		catgut.MakeFromString("project-recurse"),
	))

	m := catgut.MakeFromString("area")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		es.All.ContainsTag(m)
	}
}

func BenchmarkMatchFirstNo(b *testing.B) {
	var es Tags

	es.AddPath(MakePathWithType(
		catgut.MakeFromString("area-home"),
		catgut.MakeFromString("project-reno"),
	))

	es.AddPath(MakePathWithType(
		catgut.MakeFromString("area-career"),
		catgut.MakeFromString("project-recurse"),
	))

	m := catgut.MakeFromString("x")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		es.All.ContainsTag(m)
	}
}
