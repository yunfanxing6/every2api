package grok

import "github.com/Wei-Shaw/sub2api/internal/pkg/openai"

var DefaultModels = []openai.Model{
	{ID: "grok-4.20-0309-non-reasoning", Object: "model", Created: 1752192000, OwnedBy: "xai", Type: "model", DisplayName: "Grok 4.20 0309 Non-Reasoning (Fast)"},
	{ID: "grok-4.20-0309", Object: "model", Created: 1752192000, OwnedBy: "xai", Type: "model", DisplayName: "Grok 4.20 0309 (Auto)"},
	{ID: "grok-4.20-0309-reasoning", Object: "model", Created: 1752192000, OwnedBy: "xai", Type: "model", DisplayName: "Grok 4.20 0309 Reasoning (Expert)"},
	{ID: "grok-imagine-image-lite", Object: "model", Created: 1752192000, OwnedBy: "xai", Type: "model", DisplayName: "Grok Imagine Image Lite"},
}

func DefaultModelIDs() []string {
	ids := make([]string, len(DefaultModels))
	for i, model := range DefaultModels {
		ids[i] = model.ID
	}
	return ids
}
