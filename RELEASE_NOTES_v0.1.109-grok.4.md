# Release Notes: v0.1.109-grok.4

Maintenance release focused on CI and security workflow stability.

## Changed

- CI workflows now pin Go `1.26.2` explicitly instead of inheriting from `go.mod`
- Docker build base updated to Go `1.26.2`
- frontend audit exception dates refreshed

## Fixed

- Antigravity OAuth tests updated for environment-driven `client_id`
- API contract test updated for Grok pricing fields
- Grok default model mapping test updated for `grok-imagine-1.0-fast -> grok-imagine-1.0-fast`

## Notes

- runtime Grok functionality is unchanged from `v0.1.109-grok.3`
- this release is mainly to improve repository health and GitHub Actions signal
