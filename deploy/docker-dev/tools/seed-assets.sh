#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../.."

GO_IMAGE="${GO_IMAGE:-public.ecr.aws/docker/library/golang:1.24-alpine}"
CONFIG_PATH="${GVA_CONFIG_PATH:-../deploy/docker-dev/config.yaml}"

if ! command -v docker >/dev/null 2>&1; then
  echo "未找到 docker 命令，请在部署服务器执行。" >&2
  exit 1
fi

if [ ! -f "deploy/docker-dev/config.yaml" ]; then
  echo "缺少 deploy/docker-dev/config.yaml，请先启动或初始化 Docker 开发环境。" >&2
  exit 1
fi

docker run --rm --network host \
  -e GOPROXY="${GOPROXY:-https://goproxy.cn,direct}" \
  -e GOFLAGS="${GOFLAGS:--mod=readonly}" \
  -v "$PWD:/src" \
  -v gva-go-mod-cache:/go/pkg/mod \
  -v gva-go-build-cache:/root/.cache/go-build \
  -w /src/server \
  "$GO_IMAGE" \
  go run ./cmd/seed-assets --config "$CONFIG_PATH" "$@"
