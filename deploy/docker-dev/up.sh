#!/usr/bin/env bash
set -euo pipefail
export DOCKER_BUILDKIT=0
export COMPOSE_DOCKER_CLI_BUILD=0
cd "$(dirname "$0")"
[ -f .env ] && set -a && . ./.env && set +a
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
