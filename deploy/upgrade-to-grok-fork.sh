#!/usr/bin/env bash
set -euo pipefail

REPO_URL="https://github.com/yunfanxing6/every2api.git"
RAW_BASE="https://raw.githubusercontent.com/yunfanxing6/every2api/main/deploy"
TARGET_DIR="${TARGET_DIR:-$(pwd)}"
SOURCE_DIR="${SUB2API_FORK_SOURCE_DIR:-${TARGET_DIR}/every2api-src}"
BACKUP_DIR="${TARGET_DIR}/backup-$(date +%Y%m%d-%H%M%S)"
FORK_REF="${SUB2API_FORK_REF:-main}"

blue() { printf '\033[0;34m[INFO]\033[0m %s\n' "$1"; }
green() { printf '\033[0;32m[SUCCESS]\033[0m %s\n' "$1"; }
yellow() { printf '\033[1;33m[WARN]\033[0m %s\n' "$1"; }
red() { printf '\033[0;31m[ERROR]\033[0m %s\n' "$1"; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { red "missing command: $1"; exit 1; }
}

compose() {
  if docker compose version >/dev/null 2>&1; then
    docker compose "$@"
  elif command -v docker-compose >/dev/null 2>&1; then
    docker-compose "$@"
  else
    red "missing docker compose / docker-compose"
    exit 1
  fi
}

detect_existing_compose_file() {
  for candidate in docker-compose.grok.yml docker-compose.local.yml docker-compose.yml; do
    if [ -f "$candidate" ]; then
      printf '%s' "$candidate"
      return 0
    fi
  done
  return 1
}

ensure_env_value() {
  local key="$1"
  local value="$2"
  if grep -q "^${key}=" .env; then
    if grep -q "^${key}=$" .env; then
      perl -0pi -e "s/^${key}=.*$/${key}=${value}/m" .env
    fi
  else
    echo "${key}=${value}" >> .env
  fi
}

write_rollback_script() {
  local previous_compose="$1"
  cat > "${BACKUP_DIR}/rollback.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
cd "${TARGET_DIR}"

compose() {
  if docker compose version >/dev/null 2>&1; then
    docker compose "\$@"
  else
    docker-compose "\$@"
  fi
}

echo "[rollback] restoring backup from ${BACKUP_DIR}"
for item in .env docker-compose.yml docker-compose.local.yml docker-compose.grok.yml config.yaml .env.example.grok; do
  if [ -e "${BACKUP_DIR}/\$item" ]; then
    rm -rf "\$item"
    cp -a "${BACKUP_DIR}/\$item" "\$item"
  fi
done

for dir in data postgres_data redis_data; do
  if [ -d "${BACKUP_DIR}/\$dir" ]; then
    rm -rf "\$dir"
    cp -a "${BACKUP_DIR}/\$dir" "\$dir"
  fi
done

if [ -n "${previous_compose}" ] && [ -f "${previous_compose}" ]; then
  compose -f "${previous_compose}" up -d
fi

echo "[rollback] completed"
EOF
  chmod +x "${BACKUP_DIR}/rollback.sh"
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

  local existing_compose=""

  mkdir -p "$TARGET_DIR"
  cd "$TARGET_DIR"

  if existing_compose=$(detect_existing_compose_file); then
    blue "detected existing compose file: $existing_compose"
  else
    yellow "no existing compose file detected, migration will proceed in current directory"
    existing_compose=""
  fi

  blue "creating backup in $BACKUP_DIR"
  mkdir -p "$BACKUP_DIR"
  for item in .env docker-compose.yml docker-compose.local.yml docker-compose.grok.yml .env.example.grok config.yaml data postgres_data redis_data; do
    if [ -e "$item" ]; then
      cp -a "$item" "$BACKUP_DIR/"
    fi
  done
  write_rollback_script "$existing_compose"

  blue "downloading fork env template"
  download "$RAW_BASE/.env.example" .env.example.grok

  if [ ! -f .env ]; then
    cp .env.example.grok .env
  fi

  ensure_env_value SUB2API_RELEASE_REPO yunfanxing6/every2api
  ensure_env_value JWT_SECRET "$(openssl rand -hex 32)"
  ensure_env_value TOTP_ENCRYPTION_KEY "$(openssl rand -hex 32)"

  mkdir -p data postgres_data redis_data "$SOURCE_DIR"

  if [ -d "$SOURCE_DIR/.git" ]; then
    blue "updating local fork source checkout"
    git -C "$SOURCE_DIR" fetch --tags origin
    git -C "$SOURCE_DIR" checkout "$FORK_REF"
    git -C "$SOURCE_DIR" pull --ff-only origin "$FORK_REF" || true
  else
    blue "cloning fork source to $SOURCE_DIR"
    rm -rf "$SOURCE_DIR"
    git clone "$REPO_URL" "$SOURCE_DIR"
    git -C "$SOURCE_DIR" checkout "$FORK_REF"
  fi

  if [ -n "$existing_compose" ] && [ -f "$existing_compose" ]; then
    blue "stopping existing deployment via $existing_compose"
    compose -f "$existing_compose" down || yellow "existing stack stop failed, continuing"
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
    image: every2api-local:latest
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
  compose -f docker-compose.grok.yml build sub2api
  compose -f docker-compose.grok.yml up -d

  green "upgrade complete"
  echo
  echo "Current deployment files:"
  echo "  - $TARGET_DIR/docker-compose.grok.yml"
  echo "  - $TARGET_DIR/.env"
  echo "  - source: $SOURCE_DIR"
  echo "  - rollback: ${BACKUP_DIR}/rollback.sh"
  echo
  echo "Useful commands:"
  echo "  compose -f docker-compose.grok.yml logs -f sub2api"
  echo "  compose -f docker-compose.grok.yml pull    # only for postgres/redis"
  echo "  git -C $SOURCE_DIR fetch --tags origin && git -C $SOURCE_DIR checkout $FORK_REF && git -C $SOURCE_DIR pull --ff-only origin $FORK_REF && compose -f docker-compose.grok.yml build sub2api && compose -f docker-compose.grok.yml up -d"
  echo "  ${BACKUP_DIR}/rollback.sh"
}

main "$@"
