# Sub2API Grok Fork

<div align="center">

[![Base](https://img.shields.io/badge/Base-Sub2API%20v0.1.109-1f6feb.svg)](https://github.com/Wei-Shaw/sub2api)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3-42b883.svg)](https://vuejs.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**A Grok-enabled Sub2API fork with a first-class `grok` platform**

[中文说明](README_GROK_FORK_CN.md)

[Release Notes](RELEASE_NOTES_v0.1.109-grok.1.md)

</div>

## What This Fork Adds

This repository is a practical fork of `Sub2API` built for one goal:

- keep `Sub2API`'s user system, API keys, billing, subscriptions, dashboard, and multi-tenant gateway
- add a dedicated `grok` platform instead of hiding Grok behind generic OpenAI config
- connect `Sub2API` to `grok2api` as the Grok upstream

That means you can run a user-facing Grok gateway with:

- user registration and login
- user API keys
- per-group pricing
- Grok text, image, and video routing
- media file proxying through `Sub2API`

## Current Capabilities

Implemented in this fork:

- first-class `grok` platform in backend and frontend
- Grok group/account creation in admin UI
- default Grok model set:
  - `grok-4.20-beta`
  - `grok-imagine-1.0-fast`
  - `grok-imagine-1.0-edit`
  - `grok-imagine-1.0-video`
- Grok pricing fields at group level:
  - text input/output per MTok
  - image 1K / 2K
  - video 5s / 10s / 15s
  - HD multiplier
- OpenAI-compatible endpoints for Grok:
  - `POST /v1/chat/completions`
  - `POST /v1/messages`
  - `GET /v1/models`
- media endpoints:
  - `POST /v1/images/generations`
  - `POST /v1/images/edits`
  - `POST /v1/videos`
  - `POST /v1/video/extend`
  - `GET /v1/files/image/*`
  - `GET /v1/files/video/*`

## Verified Status

Verified working end-to-end:

- text generation
- image generation
- video generation
- media file proxying

Known limitation:

- `grok-imagine-1.0-edit`

At the moment, the remaining failure appears to be on the `grok2api` upstream adaptation side, not the `Sub2API` integration layer.

## Architecture

Recommended production layout:

1. `sub2api-grok` is your public gateway
2. `grok2api` runs behind it as the Grok upstream

Typical setup:

- public gateway: `https://sub.example.com`
- Grok upstream: `https://grok.example.com`

In the admin panel, create a Grok upstream account with:

- platform: `grok`
- type: `apikey`
- `base_url`: `https://grok.example.com/v1`
- `api_key`: your `grok2api` API key

## Why This Exists

`grok2api` is great as an upstream adapter, but by itself it is not built around:

- multi-user registration
- self-service API key issuance
- subscription/balance management
- tenant-facing admin workflows

This fork keeps those responsibilities inside `Sub2API`, and keeps `grok2api` focused on Grok upstream access.

## Security Notes

This fork includes fixes for issues encountered during integration work:

- authenticated Grok media proxy path validation
- safer media URL rewriting
- clearer upstream rate-limit error propagation

Still recommended before public release:

1. review all deployment secrets and defaults
2. verify your own pricing policy for Grok models
3. test your own `grok2api` token pool behavior under load

## Deployment Notes

This fork is intended to be deployed from source-built images, not by clicking the official Sub2API in-app updater.

Reason:

- this is not stock upstream `Sub2API`
- it contains custom Grok platform and gateway changes
- upgrading through the stock updater can overwrite fork-specific behavior

## Upgrade Paths

### Existing users of this fork

This fork's update checker now defaults to:

- `yunfanxing6/sub2api-grok`

You can override it with:

- `SUB2API_RELEASE_REPO=owner/repo`

### Existing users of stock Sub2API

Use the migration script in `deploy/upgrade-to-grok-fork.sh`.

Typical usage:

```bash
curl -sSL https://raw.githubusercontent.com/yunfanxing6/sub2api-grok/main/deploy/upgrade-to-grok-fork.sh | bash
```

What it does:

- backs up your current deployment files
- detects your existing compose file automatically
- stops the old stack before switching
- clones this fork beside your deployment
- preserves `.env`, `data`, `postgres_data`, `redis_data`
- generates a `docker-compose.grok.yml` that builds this fork from source
- sets `SUB2API_RELEASE_REPO=yunfanxing6/sub2api-grok`
- writes a rollback script into the backup directory

## Workspace

This local workspace was prepared from:

- upstream base: `Sub2API v0.1.109`

Local working directory:

- `/home/xingyunfan/sub2api-grok`

## Roadmap

Planned / worth improving:

1. finish `grok-imagine-1.0-edit`
2. reduce divergence from upstream and upstream these changes selectively
3. add dedicated fork deployment docs and release workflow

## License

This fork inherits the upstream project's licensing model. Review the upstream repository and preserve all required notices before public redistribution.
