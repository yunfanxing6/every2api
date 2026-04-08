# Release Notes: v0.1.109-grok.3

Patch release on top of `v0.1.109-grok.2`.

## Fixed

- updater version comparison now correctly handles fork-style versions like:
  - `0.1.109-grok.1`
  - `0.1.109-grok.2`
  - `0.1.109-grok.3`

This means the in-app update checker can now correctly report newer fork releases.

## Included

- all `v0.1.109-grok.2` migration script improvements
- all `v0.1.109-grok.1` Grok platform integration changes
