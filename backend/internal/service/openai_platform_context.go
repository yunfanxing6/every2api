package service

import "context"

type openAICompatiblePlatformContextKey struct{}

func WithOpenAICompatiblePlatform(ctx context.Context, platform string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if platform == "" {
		platform = PlatformOpenAI
	}
	return context.WithValue(ctx, openAICompatiblePlatformContextKey{}, platform)
}

func OpenAICompatiblePlatformFromContext(ctx context.Context) string {
	if ctx != nil {
		if platform, ok := ctx.Value(openAICompatiblePlatformContextKey{}).(string); ok && platform != "" {
			return platform
		}
	}
	return PlatformOpenAI
}
