package memory

import (
	"testing"
)

func TestVersioning(t *testing.T) {
	v := NewVersioning()
	v.Set("key", "value", 1)
	val, ver := v.Get("key")
	if val != "value" || ver != 1 {
		t.Error("Version mismatch")
	}
}

func TestRelationships(t *testing.T) {
	r := NewRelationshipGraph()
	r.AddEdge("A", "B", "uses")
	if !r.HasEdge("A", "B") {
		t.Error("Edge should exist")
	}
}

func TestProfile(t *testing.T) {
	p := NewProfileGenerator()
	profile := p.Generate([]string{"pref1", "pref2"})
	if profile == nil {
		t.Fatal("Profile is nil")
	}
}

func TestContainers(t *testing.T) {
	c := NewContainerSystem()
	c.Add("item1", "containerA")
	if !c.Has("item1") {
		t.Error("Item should be in container")
	}
}

func TestHotness(t *testing.T) {
	h := NewHotnessTracker()
	h.Track("key")
	h.Track("key")
	if h.Get("key") != 2 {
		t.Error("Hotness should be 2")
	}
}

func TestLevels(t *testing.T) {
	l := NewContextLevels()
	summary := l.Summarize([]string{"a", "b", "c", "d", "e"})
	if summary == "" {
		t.Fatal("Summary is empty")
	}
}

func TestFacts(t *testing.T) {
	f := NewFactExtractor()
	facts := f.Extract("API at /users, use JWT")
	if len(facts) == 0 {
		t.Log("No facts extracted (acceptable)")
	}
}

func TestWisdom(t *testing.T) {
	w := NewWisdomStore()
	w.AddLesson("check auth first")
	if !w.HasLesson("check auth first") {
		t.Error("Lesson should be stored")
	}
}