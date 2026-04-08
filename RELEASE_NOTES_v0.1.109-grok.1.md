# Release Notes: v0.1.109-grok.1

First Grok fork release rebased onto `Sub2API v0.1.109`.

## Added

- first-class `grok` platform in backend and frontend
- Grok admin group/account support
- Grok model presets:
  - `grok-4.20-beta`
  - `grok-imagine-1.0-fast`
  - `grok-imagine-1.0-edit`
  - `grok-imagine-1.0-video`
- Grok media routes:
  - `/v1/images/generations`
  - `/v1/images/edits`
  - `/v1/videos`
  - `/v1/video/extend`
  - `/v1/files/image/*`
  - `/v1/files/video/*`
- Grok pricing fields at group level
- one-click migration script from stock Sub2API:
  - `deploy/upgrade-to-grok-fork.sh`

## Changed

- fork updater now defaults to `yunfanxing6/sub2api-grok`
- built-in Gemini CLI / Antigravity OAuth credentials are no longer embedded in source
- Docker deployment examples expose fork release repo and built-in OAuth env variables explicitly
- image edit path in `grok2api` is now fixed in the deployed upstream stack

## Synced From Upstream v0.1.109

- updated base version to `Sub2API v0.1.109`
- kept Grok-specific OpenAI normalization safeguards so Grok models are not rewritten into OpenAI Codex defaults
- preserved fork updater compatibility with public GitHub releases

## Security

- hardened Grok media file path validation
- safer media URL rewriting behavior
- clearer upstream rate-limit error propagation
- removed embedded third-party OAuth credential values from the public repository

## Verified

- text generation
- image generation
- image edit
- video generation
- media file proxying
- in-app update check against fork releases
