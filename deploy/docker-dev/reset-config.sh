#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
cp config.yaml "config.yaml.bak.$(date +%Y%m%d%H%M%S)" 2>/dev/null || true
cp config.init.yaml config.yaml
chmod 600 config.yaml
if grep -q '^GVA_RUSTFS_ENDPOINT=' .env 2>/dev/null; then
  ./configure-rustfs.sh
fi
echo "已重置 config.yaml；不会删除 PostgreSQL 中已有数据库。"
