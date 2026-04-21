package planning

import (
	"context"

	"github.com/hermes-v2/aigo/internal/providers"
)

type PlanningLLMProvider struct {
	pm           *providers.ProviderManager
	providerName string
}

func NewPlanningLLMProvider(pm *providers.ProviderManager, providerName string) *PlanningLLMProvider {
	return &PlanningLLMProvider{
		pm:           pm,
		providerName: providerName,
	}
}

func (p *PlanningLLMProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	provider, err := p.pm.Get(p.providerName)
	if err != nil {
		return "", err
	}

	providerMsgs := make([]providers.Message, len(messages))
	for i, m := range messages {
		providerMsgs[i] = providers.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	resp, err := provider.Chat(ctx, providerMsgs, nil)
	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", nil
	}

	return resp.Content, nil
}