#!/usr/bin/env bash
set -euo pipefail
export DOCKER_BUILDKIT=0
export COMPOSE_DOCKER_CLI_BUILD=0
cd "$(dirname "$0")"
if [ ! -f .env ]; then
  echo "缺少 deploy/docker-dev/.env，请先执行：cp .env.example .env" >&2
  exit 1
fi
set -a
. ./.env
set +a

required_vars=(
  GVA_PG_HOST GVA_PG_PORT GVA_PG_USER GVA_PG_PASSWORD GVA_PG_DB
  GVA_ADMIN_PASSWORD GVA_REDIS_ADDR GVA_RUSTFS_ENDPOINT
  GVA_RUSTFS_ACCESS_KEY GVA_RUSTFS_SECRET_KEY GVA_RUSTFS_BUCKET
)
for name in "${required_vars[@]}"; do
  if [ -z "${!name:-}" ]; then
    echo ".env 缺少必填变量：${name}" >&2
    exit 1
  fi
done

if [ ! -f config.yaml ]; then
  cp config.init.yaml config.yaml
  chmod 600 config.yaml
fi
if [ -n "${GVA_RUSTFS_ENDPOINT:-}" ]; then
  ./configure-rustfs.sh
fi
if docker image inspect registry.cn-zhangjiakou.aliyuncs.com/yunli_mid_platform/nginx:alpine >/dev/null 2>&1; then
  docker tag registry.cn-zhangjiakou.aliyuncs.com/yunli_mid_platform/nginx:alpine nginx:alpine >/dev/null
fi
docker compose --env-file .env -f docker-compose.yml up -d --build --force-recreate
./init-db.sh
docker compose --env-file .env -f docker-compose.yml ps
./health-check.sh
