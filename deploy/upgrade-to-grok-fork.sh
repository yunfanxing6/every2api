#!/usr/bin/env bash
set -euo pipefail

REPO_URL="https://github.com/yunfanxing6/sub2api-grok.git"
RAW_BASE="https://raw.githubusercontent.com/yunfanxing6/sub2api-grok/main/deploy"
TARGET_DIR="${TARGET_DIR:-$(pwd)}"
SOURCE_DIR="${SUB2API_FORK_SOURCE_DIR:-${TARGET_DIR}/sub2api-grok-src}"
BACKUP_DIR="${TARGET_DIR}/backup-$(date +%Y%m%d-%H%M%S)"

blue() { printf '\033[0;34m[INFO]\033[0m %s\n' "$1"; }
green() { printf '\033[0;32m[SUCCESS]\033[0m %s\n' "$1"; }
yellow() { printf '\033[1;33m[WARN]\033[0m %s\n' "$1"; }
red() { printf '\033[0;31m[ERROR]\033[0m %s\n' "$1"; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { red "missing command: $1"; exit 1; }
}

download() {
  local url="$1"
  local dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$dest"
  else
    wget -q "$url" -O "$dest"
  fi
}

main() {
  require_cmd git
  require_cmd docker
  require_cmd openssl

  mkdir -p "$TARGET_DIR"
  cd "$TARGET_DIR"

  blue "creating backup in $BACKUP_DIR"
  mkdir -p "$BACKUP_DIR"
  for item in .env docker-compose.yml docker-compose.local.yml config.yaml data postgres_data redis_data; do
    if [ -e "$item" ]; then
      cp -a "$item" "$BACKUP_DIR/"
    fi
  done

  blue "downloading fork env template"
  download "$RAW_BASE/.env.example" .env.example.grok

  if [ ! -f .env ]; then
    cp .env.example.grok .env
  fi

  if ! grep -q '^SUB2API_RELEASE_REPO=' .env; then
    echo 'SUB2API_RELEASE_REPO=yunfanxing6/sub2api-grok' >> .env
  else
    perl -0pi -e 's/^SUB2API_RELEASE_REPO=.*/SUB2API_RELEASE_REPO=yunfanxing6\/sub2api-grok/m' .env
  fi

  if ! grep -q '^JWT_SECRET=' .env; then
    echo "JWT_SECRET=$(openssl rand -hex 32)" >> .env
  fi
  if ! grep -q '^TOTP_ENCRYPTION_KEY=' .env; then
    echo "TOTP_ENCRYPTION_KEY=$(openssl rand -hex 32)" >> .env
  fi

  mkdir -p data postgres_data redis_data "$SOURCE_DIR"

  if [ -d "$SOURCE_DIR/.git" ]; then
    blue "updating local fork source checkout"
    git -C "$SOURCE_DIR" fetch --tags origin
    git -C "$SOURCE_DIR" checkout main
    git -C "$SOURCE_DIR" pull --ff-only origin main
  else
    blue "cloning fork source to $SOURCE_DIR"
    rm -rf "$SOURCE_DIR"
    git clone "$REPO_URL" "$SOURCE_DIR"
  fi

  blue "generating docker-compose.grok.yml"
  cat > docker-compose.grok.yml <<EOF
services:
  sub2api:
    build:
      context: ${SOURCE_DIR}
      dockerfile: deploy/Dockerfile
      args:
        GOPROXY: \\${GOPROXY:-https://goproxy.cn,direct}
        GOSUMDB: \\${GOSUMDB:-sum.golang.google.cn}
    image: sub2api-grok-local:latest
    container_name: sub2api
    restart: unless-stopped
    ulimits:
      nofile:
        soft: 100000
        hard: 100000
    ports:
      - "\\${BIND_HOST:-0.0.0.0}:\\${SERVER_PORT:-8080}:8080"
    volumes:
      - ./data:/app/data
    env_file:
      - .env
    environment:
      - AUTO_SETUP=true
      - SERVER_HOST=0.0.0.0
      - SERVER_PORT=8080
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  postgres:
    image: postgres:18-alpine
    container_name: sub2api-postgres
    restart: unless-stopped
    volumes:
      - ./postgres_data:/var/lib/postgresql/data
    env_file:
      - .env
    environment:
      - PGDATA=/var/lib/postgresql/data
      - POSTGRES_USER=\\${POSTGRES_USER:-sub2api}
      - POSTGRES_PASSWORD=\\${POSTGRES_PASSWORD}
      - POSTGRES_DB=\\${POSTGRES_DB:-sub2api}
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U \\${POSTGRES_USER:-sub2api} -d \\${POSTGRES_DB:-sub2api}"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:8-alpine
    container_name: sub2api-redis
    restart: unless-stopped
    volumes:
      - ./redis_data:/data
    env_file:
      - .env
    command: >
      sh -c 'redis-server --save 60 1 --appendonly yes --appendfsync everysec \\${REDIS_PASSWORD:+--requirepass "\\$REDIS_PASSWORD"}'
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
EOF

  blue "building and starting Grok fork"
  docker compose -f docker-compose.grok.yml build sub2api
  docker compose -f docker-compose.grok.yml up -d

  green "upgrade complete"
  echo
  echo "Current deployment files:"
  echo "  - $TARGET_DIR/docker-compose.grok.yml"
  echo "  - $TARGET_DIR/.env"
  echo "  - source: $SOURCE_DIR"
  echo
  echo "Useful commands:"
  echo "  docker compose -f docker-compose.grok.yml logs -f sub2api"
  echo "  docker compose -f docker-compose.grok.yml pull    # only for postgres/redis"
  echo "  git -C $SOURCE_DIR pull --ff-only && docker compose -f docker-compose.grok.yml build sub2api && docker compose -f docker-compose.grok.yml up -d"
}

main "$@"
