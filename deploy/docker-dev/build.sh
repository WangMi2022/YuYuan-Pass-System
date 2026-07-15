#!/usr/bin/env bash
set -euo pipefail
export DOCKER_BUILDKIT=0
export COMPOSE_DOCKER_CLI_BUILD=0
cd "$(dirname "$0")"
if docker image inspect registry.cn-zhangjiakou.aliyuncs.com/yunli_mid_platform/nginx:alpine >/dev/null 2>&1; then
  docker tag registry.cn-zhangjiakou.aliyuncs.com/yunli_mid_platform/nginx:alpine nginx:alpine >/dev/null
fi
docker compose --env-file .env -f docker-compose.yml build "$@"
