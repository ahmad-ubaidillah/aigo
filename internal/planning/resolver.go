package planning

import (
	"context"
)

type ResolvedGap struct {
	Original   *Gap
	Resolution string
	Applied   bool
}

type Resolver struct {
	defaults map[string]string
}

func NewResolver() *Resolver {
	return &Resolver{
		defaults: map[string]string{
			"ambiguous":  "clarify requirement",
			"empty_steps": "add explicit steps",
			"missing_plan": "create plan first",
		},
	}
}

func (r *Resolver) ResolveGap(ctx context.Context, gap *Gap) *ResolvedGap {
	if gap == nil {
		return &ResolvedGap{
			Original:   nil,
			Resolution: "no gap to resolve",
			Applied:    false,
		}
	}

	resolution, exists := r.defaults[gap.Type]
	if !exists {
		resolution = "manual review required"
	}

	return &ResolvedGap{
		Original:   gap,
		Resolution: resolution,
		Applied:    true,
	}
}

func (r *Resolver) ResolveAmbiguity(input string) string {
	ambiguousPhrases := []string{"do it", "fix it", "something"}

	for _, phrase := range ambiguousPhrases {
		if input == phrase {
			return "requirement unclear: " + input
		}
	}

	return input
}