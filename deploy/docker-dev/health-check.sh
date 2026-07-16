#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

if [ ! -f .env ]; then
  echo "缺少 deploy/docker-dev/.env，请先执行：cp .env.example .env" >&2
  exit 1
fi

set -a
. ./.env
set +a

WEB_URL="http://127.0.0.1:${WEB_PORT:-8080}"
SERVER_URL="http://127.0.0.1:${SERVER_PORT:-8888}"

check_url() {
  local name="$1"
  local method="$2"
  local url="$3"
  local response
  local attempt

  for attempt in $(seq 1 15); do
    if [ "$method" = "POST" ]; then
      if response=$(curl -fsS --max-time 5 -X POST "$url" 2>/dev/null); then
        echo "[OK] ${name}: ${url}"
        printf '%s' "$response"
        return 0
      fi
    elif response=$(curl -fsS --max-time 5 "$url" 2>/dev/null); then
      echo "[OK] ${name}: ${url}"
      printf '%s' "$response"
      return 0
    fi
    sleep 2
  done

  echo "[FAIL] ${name}: ${url} 在 30 秒内未就绪" >&2
  return 1
}

echo "检查容器状态"
docker compose --env-file .env -f docker-compose.yml ps

echo "检查 Web 服务"
check_url "Web" GET "${WEB_URL}/" >/dev/null
echo "[OK] Web: ${WEB_URL}/"

echo "检查 API 与数据库状态"
DB_RESPONSE=$(check_url "API" POST "${SERVER_URL}/init/checkdb")
if ! printf '%s' "$DB_RESPONSE" | grep -q '"needInit":false'; then
  echo "[FAIL] API 可访问，但数据库尚未完成初始化" >&2
  exit 1
fi
echo "[OK] API 数据库已初始化"

echo "健康检查通过"
