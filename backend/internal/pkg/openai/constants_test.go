package openai

import "testing"

func TestDefaultModelIDsExposeCurrentOpenAIFamily(t *testing.T) {
	ids := DefaultModelIDs()
	seen := make(map[string]struct{}, len(ids))

	for _, id := range ids {
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate default model id: %s", id)
		}
		seen[id] = struct{}{}
	}

	expected := []string{
		"gpt-5.5",
		"gpt-5.4",
		"gpt-5.4-mini",
		"gpt-5.4-2026-03-05",
		"gpt-5.3-codex",
		"gpt-5.3-codex-spark",
		"gpt-5.2",
		"gpt-5.2-2025-12-11",
		"gpt-5.2-chat-latest",
		"gpt-5.2-pro",
		"gpt-5.2-pro-2025-12-11",
		"gpt-image-1",
		"gpt-image-1.5",
		"gpt-image-2",
	}

	for _, id := range expected {
		if _, ok := seen[id]; !ok {
			t.Fatalf("default model id %q missing from OpenAI defaults", id)
		}
	}
}
