package permission

import (
	"testing"
)

func TestRuleset_DefaultAllow(t *testing.T) {
	t.Parallel()

	r := NewRuleset()
	if r.Check("bash") != PermAllow {
		t.Error("expected default allow")
	}
}

func TestRuleset_Deny(t *testing.T) {
	t.Parallel()

	r := NewRuleset()
	r.Deny("bash")
	if r.Check("bash") != PermDeny {
		t.Error("expected deny")
	}
}

func TestRuleset_Ask(t *testing.T) {
	t.Parallel()

	r := NewRuleset()
	r.Ask("edit")
	if r.Check("edit") != PermAsk {
		t.Error("expected ask")
	}
}

func TestRuleset_Allow(t *testing.T) {
	t.Parallel()

	r := NewRuleset()
	r.Allow("read")
	if r.Check("read") != PermAllow {
		t.Error("expected allow")
	}
}

func TestRuleset_Wildcard(t *testing.T) {
	t.Parallel()

	r := NewRuleset()
	r.AddRule("web*", "*", PermDeny)
	if r.Check("webfetch") != PermDeny {
		t.Error("expected wildcard deny")
	}
	if r.Check("bash") != PermAllow {
		t.Error("expected allow for non-matching")
	}
}

func TestRuleset_Clear(t *testing.T) {
	t.Parallel()

	r := NewRuleset()
	r.Deny("bash")
	r.Clear()
	if r.Check("bash") != PermAllow {
		t.Error("expected allow after clear")
	}
}

func TestRuleset_Rules(t *testing.T) {
	t.Parallel()

	r := NewRuleset()
	r.Deny("bash")
	r.Ask("edit")
	rules := r.Rules()
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}

func TestRuleset_ParseLevel(t *testing.T) {
	t.Parallel()

	r := NewRuleset()

	level, err := r.ParseLevel("allow")
	if err != nil || level != PermAllow {
		t.Errorf("expected allow, got %v, %v", level, err)
	}

	level, err = r.ParseLevel("deny")
	if err != nil || level != PermDeny {
		t.Errorf("expected deny, got %v, %v", level, err)
	}

	level, err = r.ParseLevel("ask")
	if err != nil || level != PermAsk {
		t.Errorf("expected ask, got %v, %v", level, err)
	}

	_, err = r.ParseLevel("invalid")
	if err == nil {
		t.Error("expected error for invalid level")
	}
}

func TestPermissionLevel_String(t *testing.T) {
	t.Parallel()

	if PermAllow.String() != "allow" {
		t.Errorf("expected allow, got %s", PermAllow.String())
	}
	if PermAsk.String() != "ask" {
		t.Errorf("expected ask, got %s", PermAsk.String())
	}
	if PermDeny.String() != "deny" {
		t.Errorf("expected deny, got %s", PermDeny.String())
	}
}

func TestPermissionLevel_String_Unknown(t *testing.T) {
	t.Parallel()

	level := PermissionLevel(99)
	if level.String() != "unknown" {
		t.Errorf("expected unknown, got %s", level.String())
	}
}
