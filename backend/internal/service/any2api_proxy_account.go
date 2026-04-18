package service

import (
	"context"
	"fmt"
	"strings"
)

const any2APIProxyAccountExtraKey = "system_upstream_proxy"

func any2APIProxyAccountSpec(model string) (name string, platform string, marker string) {
	platform = PlatformOpenAI
	marker = "any2api:openai"
	name = "any2api-upstream"
	lowered := strings.ToLower(strings.TrimSpace(model))
	if strings.HasPrefix(lowered, "grok") {
		platform = PlatformGrok
		marker = "any2api:grok"
		name = "grok2api-upstream"
	} else if strings.HasPrefix(lowered, "qwen") {
		platform = PlatformQwen
		marker = "any2api:qwen"
		name = "qwen2api-upstream"
	}
	return name, platform, marker
}

func matchesAny2APIProxyAccount(account Account, platform string, marker string) bool {
	if !strings.EqualFold(strings.TrimSpace(account.Platform), platform) {
		return false
	}
	if account.Extra == nil {
		return false
	}
	value, ok := account.Extra[any2APIProxyAccountExtraKey]
	if !ok || value == nil {
		return false
	}
	return strings.TrimSpace(fmt.Sprint(value)) == marker
}

// EnsureAny2APIProxyAccount returns a persisted placeholder account for Any2API usage logs.
// These accounts are usage-only and must never participate in scheduling.
func (s *OpenAIGatewayService) EnsureAny2APIProxyAccount(ctx context.Context, model string) (*Account, error) {
	if s == nil || s.accountRepo == nil {
		return nil, fmt.Errorf("account repository is unavailable")
	}

	name, platform, marker := any2APIProxyAccountSpec(model)
	accounts, err := s.accountRepo.FindByExtraField(ctx, any2APIProxyAccountExtraKey, marker)
	if err != nil {
		return nil, err
	}
	for i := range accounts {
		if matchesAny2APIProxyAccount(accounts[i], platform, marker) {
			account := accounts[i]
			return &account, nil
		}
	}

	account := &Account{
		Name:        name,
		Platform:    platform,
		Type:        AccountTypeAPIKey,
		Credentials: map[string]any{},
		Extra: map[string]any{
			any2APIProxyAccountExtraKey: marker,
		},
		Concurrency: 1,
		Priority:    0,
		Status:      StatusActive,
		Schedulable: false,
	}
	if err := s.accountRepo.Create(ctx, account); err != nil {
		accounts, lookupErr := s.accountRepo.FindByExtraField(ctx, any2APIProxyAccountExtraKey, marker)
		if lookupErr == nil {
			for i := range accounts {
				if matchesAny2APIProxyAccount(accounts[i], platform, marker) {
					existing := accounts[i]
					return &existing, nil
				}
			}
		}
		return nil, err
	}
	return account, nil
}
