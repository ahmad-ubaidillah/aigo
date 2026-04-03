package skills

import (
	"testing"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

func TestRegistry_Register(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	err := r.Register("test", "Test skill", "echo test", "general", "test")
	if err != nil {
		t.Fatal(err)
	}
	skills, _ := r.List("general")
	if len(skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(skills))
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register("test", "Test skill", "echo test", "general", "test")
	err := r.Register("test", "Duplicate", "echo dup", "general", "dup")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegistry_LoadSkill(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register("test", "Test skill", "echo test", "general", "test")
	skill, err := r.LoadSkill("test")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name != "test" {
		t.Errorf("expected test, got %s", skill.Name)
	}
}

func TestRegistry_LoadSkillNotFound(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	_, err := r.LoadSkill("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRegistry_List(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register("a", "Skill A", "echo a", "general", "a")
	r.Register("b", "Skill B", "echo b", "coding", "b")
	all, _ := r.List("")
	if len(all) != 2 {
		t.Errorf("expected 2 skills, got %d", len(all))
	}
	gen, _ := r.List("general")
	if len(gen) != 1 {
		t.Errorf("expected 1 general skill, got %d", len(gen))
	}
}

func TestRegistry_Execute(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register("test", "Test skill", "echo test", "general", "test")
	result, err := r.Execute("test", "args")
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected result")
	}
}

func TestRegistry_ExecuteNotFound(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	_, err := r.Execute("nonexistent", "args")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRegistry_ExecuteDisabled(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register("test", "Test skill", "echo test", "general", "test")
	r.skills["test"].Enabled = false
	_, err := r.Execute("test", "args")
	if err == nil {
		t.Error("expected error for disabled skill")
	}
}

func TestRegistry_ExecuteRaw(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register("test", "Test skill", "echo test", "general", "test")
	result, err := r.ExecuteRaw("echo hello")
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected result")
	}
}

func TestRegistry_Search(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register("search-test", "A searchable skill", "echo test", "general", "search tag")
	results, err := r.Search("searchable")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestRegistry_GetByCategory(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register("a", "Skill A", "echo a", "general", "a")
	r.Register("b", "Skill B", "echo b", "coding", "b")
	results := r.GetByCategory("general")
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestRegistry_SaveSkill(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	skill := &types.Skill{ID: "s1", Name: "saved", Category: "general", Enabled: true}
	err := r.SaveSkill(skill)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegistry_SaveNilSkill(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	err := r.SaveSkill(nil)
	if err == nil {
		t.Error("expected error for nil skill")
	}
}

func TestRegistry_DeleteSkill(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register("test", "Test skill", "echo test", "general", "test")
	err := r.DeleteSkill("test")
	if err != nil {
		t.Fatal(err)
	}
	all, _ := r.List("")
	if len(all) != 0 {
		t.Error("expected 0 skills after delete")
	}
}

func TestRegistry_DeleteSkillNotFound(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	err := r.DeleteSkill("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRegistry_SetWorkspace(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.SetWorkspace("/tmp")
}

func TestRegistry_AddBuiltIn(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.AddBuiltIn(nil)
}

func TestNewRegistryWithWorkspace(t *testing.T) {
	t.Parallel()
	r := NewRegistryWithWorkspace("/tmp")
	if r == nil {
		t.Error("expected registry")
	}
}

func TestExecutor(t *testing.T) {
	t.Parallel()
	e := NewExecutor("", 60*time.Second)
	if e == nil {
		t.Error("expected executor")
	}
}

func TestMarketplace(t *testing.T) {
	t.Parallel()
	m := NewMarketplace(nil)
	if m == nil {
		t.Error("expected marketplace")
	}
}

func TestSkillEngine(t *testing.T) {
	t.Parallel()
	e := NewEngine(nil)
	if e == nil {
		t.Error("expected engine")
	}
}
