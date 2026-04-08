# Release Notes: v0.1.109-grok.2

Maintenance release on top of `v0.1.109-grok.1`.

## Changed

- improved `deploy/upgrade-to-grok-fork.sh`
  - auto-detect existing compose file
  - stop old stack before switching
  - preserve current deployment data and env files
  - write a rollback script into the backup directory
  - support selecting a fork branch/tag via `SUB2API_FORK_REF`

## Docs

- updated root README and Chinese README with stronger migration notes
- updated deployment README to explain the safer migration flow

## Recommended For

- users migrating from stock `Sub2API`
- existing users who want a safer, documented migration path for future hosts
