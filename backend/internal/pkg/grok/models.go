package grok

import "github.com/Wei-Shaw/sub2api/internal/pkg/openai"

var DefaultModels = []openai.Model{
	{ID: "grok-4.20-beta", Object: "model", Created: 1752192000, OwnedBy: "xai", Type: "model", DisplayName: "Grok 4.20 Beta"},
	{ID: "grok-imagine-1.0-fast", Object: "model", Created: 1752192000, OwnedBy: "xai", Type: "model", DisplayName: "Grok Imagine Fast"},
	{ID: "grok-imagine-1.0-edit", Object: "model", Created: 1752192000, OwnedBy: "xai", Type: "model", DisplayName: "Grok Imagine Edit"},
	{ID: "grok-imagine-1.0-video", Object: "model", Created: 1752192000, OwnedBy: "xai", Type: "model", DisplayName: "Grok Imagine Video"},
}

func DefaultModelIDs() []string {
	ids := make([]string, len(DefaultModels))
	for i, model := range DefaultModels {
		ids[i] = model.ID
	}
	return ids
}
