# Release Notes: v0.1.110-grok.1

## Base Upgrade

- rebased the fork feature set onto upstream `Sub2API v0.1.110`
- pulled in upstream fixes for channel service, OAuth refresh handling, Responses compatibility, and gateway request forwarding settings

## Grok Fork Compatibility

- preserved first-class `grok` platform support across backend, routing, billing, and admin UI
- kept fork updater defaults pointing to `yunfanxing6/sub2api-grok`
- retained Grok media routing and file proxy support while aligning OpenAI/Anthropic gateway behavior with upstream `v0.1.110`

## Gateway Updates

- added `enable_cch_signing` admin setting from upstream `v0.1.110`
- synced Anthropic billing header handling with upstream CCH signing and `cc_version` behavior
- ported OpenAI content-based session seed fallback and empty base64 input image sanitization
- ported non-streaming SSE-to-JSON fallback handling for OpenAI-compatible upstreams

## Verification

- backend: `go test ./...`
- frontend: `corepack pnpm run build`
- frontend: `corepack pnpm run lint:check`
