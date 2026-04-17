# Release Notes: v0.1.113-grok.1

## Base Upgrade

- rebased the fork feature set onto upstream `Sub2API v0.1.113`
- pulled in upstream `v0.1.111`-`v0.1.113` changes for admin settings, payment flow, usage/account cost views, web search support, and OpenAI messages dispatch model config

## Fork Compatibility

- retained multi-group API key routing and same-platform precise group matching
- restored `grok` / `qwen` / `any2api` integration paths after the upstream sync
- kept updater defaults pointing to `yunfanxing6/every2api`

## Verification

- backend: `go test ./...`
- backend: `go test -tags unit ./internal/server/middleware/...`
- frontend: `corepack pnpm run typecheck`
- frontend: `corepack pnpm run build`
