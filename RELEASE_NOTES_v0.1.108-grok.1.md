# Release Notes: v0.1.108-grok.1

First public Grok fork release based on `Sub2API v0.1.108`.

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
- media URL rewriting and authenticated media proxying

## Changed

- updater now defaults to `yunfanxing6/sub2api-grok`
- built-in Gemini CLI / Antigravity OAuth credentials are no longer embedded in source
- Docker deployment examples now expose fork update repo and built-in OAuth env variables explicitly

## Security

- hardened Grok media file path validation
- safer media URL rewriting behavior
- clearer upstream rate-limit error propagation
- removed embedded third-party OAuth credential values from the public repository

## Known Limitation

- `grok-imagine-1.0-edit` is still not fully stable end-to-end

## Upgrade Notes

- existing users of this fork can use the in-app updater against this repo
- stock `Sub2API` users should use `deploy/upgrade-to-grok-fork.sh` for migration
