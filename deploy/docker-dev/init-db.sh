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

: "${MAA_DB_TYPE:?请在 .env 设置 MAA_DB_TYPE}"
: "${MAA_PG_HOST:?请在 .env 设置 MAA_PG_HOST}"
: "${MAA_PG_PORT:?请在 .env 设置 MAA_PG_PORT}"
: "${MAA_PG_USER:?请在 .env 设置 MAA_PG_USER}"
: "${MAA_PG_PASSWORD:?请在 .env 设置 MAA_PG_PASSWORD}"
: "${MAA_PG_DB:?请在 .env 设置 MAA_PG_DB}"
: "${MAA_ADMIN_PASSWORD:?请在 .env 设置 MAA_ADMIN_PASSWORD}"

SERVER_URL="http://127.0.0.1:${SERVER_PORT:-8888}"
echo "等待后端启动：${SERVER_URL}"
for i in $(seq 1 90); do
  if RESP=$(curl -fsS -X POST "${SERVER_URL}/init/checkdb" 2>/dev/null); then
    break
  fi
  sleep 2
  if [ "$i" = "90" ]; then
    echo "后端启动超时，请查看日志：./logs.sh server" >&2
    exit 1
  fi
done

echo "checkdb: ${RESP}"
if echo "$RESP" | grep -q '"needInit":false'; then
  echo "数据库已初始化，跳过 initdb。"
  exit 0
fi

TMP_JSON=$(mktemp)
cat > "$TMP_JSON" <<JSON
{
  "adminPassword": "${MAA_ADMIN_PASSWORD}",
  "dbType": "${MAA_DB_TYPE}",
  "host": "${MAA_PG_HOST}",
  "port": "${MAA_PG_PORT}",
  "userName": "${MAA_PG_USER}",
  "password": "${MAA_PG_PASSWORD}",
  "dbName": "${MAA_PG_DB}",
  "template": "${MAA_PG_TEMPLATE}"
}
JSON

INIT_RESP=$(curl -fsS -X POST "${SERVER_URL}/init/initdb" \
  -H 'Content-Type: application/json' \
  --data-binary "@$TMP_JSON" || true)
rm -f "$TMP_JSON"

echo "initdb: ${INIT_RESP}"
if echo "$INIT_RESP" | grep -q '"code":0'; then
  echo "数据库初始化完成。"
  exit 0
fi

echo "数据库初始化未成功，请查看后端日志：./logs.sh server" >&2
exit 1
