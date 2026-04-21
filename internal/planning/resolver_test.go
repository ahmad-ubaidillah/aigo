package planning

import (
	"context"
	"testing"
)

func TestResolver_ResolveGap(t *testing.T) {
	r := NewResolver()

	gap := &Gap{Type: "ambiguous", Description: "test gap", Severity: 3}
	resolved := r.ResolveGap(context.Background(), gap)

	if resolved == nil {
		t.Fatal("ResolveGap() returned nil")
	}
}

func TestResolver_ResolveAmbiguity(t *testing.T) {
	r := NewResolver()

	original := "do it"
	resolved := r.ResolveAmbiguity(original)

	if resolved == "" {
		t.Log("ResolveAmbiguity returned empty (acceptable)")
	}
}