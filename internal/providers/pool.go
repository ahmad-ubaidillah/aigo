// Package providers — Provider failover with account pool.
// Inspired by NekoClaw's account pool with health-based selection.
package providers

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Account represents a single API account/key.
type Account struct {
	Name      string
	Provider  Provider
	Healthy   bool
	Cooldown  time.Time
	FailCount int
	LastUsed  time.Time
}

// AccountPool manages multiple accounts for failover.
type Pool struct {
	accounts []*Account
	mu       sync.RWMutex
	index    int // round-robin index
}

// NewPool creates a new account pool.
func NewPool() *Pool {
	return &Pool{}
}

// Add adds an account to the pool.
func (p *Pool) Add(name string, provider Provider) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.accounts = append(p.accounts, &Account{
		Name:     name,
		Provider: provider,
		Healthy:  true,
	})
	log.Printf("Account pool: added %s", name)
}

// Get returns a healthy account using round-robin with failover.
func (p *Pool) Get() (*Account, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.accounts) == 0 {
		return nil, fmt.Errorf("no accounts in pool")
	}

	now := time.Now()

	// Try round-robin first
	for i := 0; i < len(p.accounts); i++ {
		p.index = (p.index + 1) % len(p.accounts)
		acct := p.accounts[p.index]

		// Check if cooldown expired
		if !acct.Healthy && now.After(acct.Cooldown) {
			acct.Healthy = true
			acct.FailCount = 0
			log.Printf("Account pool: %s recovered from cooldown", acct.Name)
		}

		if acct.Healthy {
			acct.LastUsed = now
			return acct, nil
		}
	}

	// All accounts unhealthy — reset and try first
	log.Println("Account pool: all accounts unhealthy, resetting first account")
	acct := p.accounts[0]
	acct.Healthy = true
	acct.FailCount = 0
	acct.LastUsed = now
	return acct, nil
}

// MarkFailure marks an account as failed with exponential backoff.
func (p *Pool) MarkFailure(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, acct := range p.accounts {
		if acct.Name == name {
			acct.FailCount++
			// Exponential backoff: 10s, 30s, 90s, 270s, max 15min
			backoff := time.Duration(10) * time.Second
			for i := 1; i < acct.FailCount && i < 5; i++ {
				backoff *= 3
			}
			if backoff > 15*time.Minute {
				backoff = 15 * time.Minute
			}
			acct.Cooldown = time.Now().Add(backoff)
			acct.Healthy = false
			log.Printf("Account pool: %s failed (%d times), cooldown %s", name, acct.FailCount, backoff)
			return
		}
	}
}

// MarkSuccess resets failure count for an account.
func (p *Pool) MarkSuccess(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, acct := range p.accounts {
		if acct.Name == name {
			acct.FailCount = 0
			acct.Healthy = true
			return
		}
	}
}

// Stats returns pool statistics.
func (p *Pool) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	healthy := 0
	accounts := make([]map[string]interface{}, 0, len(p.accounts))
	for _, acct := range p.accounts {
		if acct.Healthy {
			healthy++
		}
		accounts = append(accounts, map[string]interface{}{
			"name":      acct.Name,
			"healthy":   acct.Healthy,
			"failCount": acct.FailCount,
		})
	}
	return map[string]interface{}{
		"total":    len(p.accounts),
		"healthy":  healthy,
		"accounts": accounts,
	}
}

// PooledProvider wraps a Pool to implement the Provider interface.
type PooledProvider struct {
	pool *Pool
	name string
}

// NewPooledProvider creates a provider backed by an account pool.
func NewPooledProvider(name string, pool *Pool) *PooledProvider {
	return &PooledProvider{name: name, pool: pool}
}

func (p *PooledProvider) Name() string    { return p.name }
func (p *PooledProvider) GetModel() string { return "pooled" }

func (p *PooledProvider) Chat(ctx context.Context, messages []Message, tools []ToolDef) (*Response, error) {
	acct, err := p.pool.Get()
	if err != nil {
		return nil, err
	}

	resp, err := acct.Provider.Chat(ctx, messages, tools)
	if err != nil {
		p.pool.MarkFailure(acct.Name)
		// Try next account
		acct2, err2 := p.pool.Get()
		if err2 != nil {
			return nil, err
		}
		return acct2.Provider.Chat(ctx, messages, tools)
	}

	p.pool.MarkSuccess(acct.Name)
	return resp, nil
}
