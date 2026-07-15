#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
set -a
. ./.env
set +a

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
  "adminPassword": "${GVA_ADMIN_PASSWORD}",
  "dbType": "${GVA_DB_TYPE}",
  "host": "${GVA_PG_HOST}",
  "port": "${GVA_PG_PORT}",
  "userName": "${GVA_PG_USER}",
  "password": "${GVA_PG_PASSWORD}",
  "dbName": "${GVA_PG_DB}",
  "template": "${GVA_PG_TEMPLATE}"
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
