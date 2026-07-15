#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

[ -f .env ] && set -a && . ./.env && set +a

: "${GVA_RUSTFS_ENDPOINT:?请在 .env 设置 GVA_RUSTFS_ENDPOINT（S3 API 地址，不含 http://）}"
: "${GVA_RUSTFS_ACCESS_KEY:?请在 .env 设置 GVA_RUSTFS_ACCESS_KEY}"
: "${GVA_RUSTFS_SECRET_KEY:?请在 .env 设置 GVA_RUSTFS_SECRET_KEY}"

export GVA_RUSTFS_BUCKET="${GVA_RUSTFS_BUCKET:-gva-assets}"
export GVA_RUSTFS_BASE_PATH="${GVA_RUSTFS_BASE_PATH:-assets}"
export GVA_RUSTFS_USE_SSL="${GVA_RUSTFS_USE_SSL:-false}"

python3 - <<'PY'
import json
import os
import re
from pathlib import Path

endpoint = os.environ["GVA_RUSTFS_ENDPOINT"].strip()
for prefix in ("http://", "https://"):
    if endpoint.startswith(prefix):
        endpoint = endpoint[len(prefix):]
endpoint = endpoint.rstrip("/")
access_key = os.environ["GVA_RUSTFS_ACCESS_KEY"]
secret_key = os.environ["GVA_RUSTFS_SECRET_KEY"]
bucket = os.environ["GVA_RUSTFS_BUCKET"]
base_path = os.environ["GVA_RUSTFS_BASE_PATH"].strip("/")
use_ssl = os.environ["GVA_RUSTFS_USE_SSL"].lower() in {"1", "true", "yes", "on"}
redis_addr = os.environ.get("GVA_REDIS_ADDR", "127.0.0.1:6379").strip()
redis_password = os.environ.get("GVA_REDIS_PASSWORD", "")
redis_db = int(os.environ.get("GVA_REDIS_DB", "0"))
use_redis = os.environ.get("GVA_USE_REDIS", "true").lower() in {"1", "true", "yes", "on"}
scheme = "https" if use_ssl else "http"

def q(value):
    return json.dumps(value, ensure_ascii=False)

block = f'''minio:
    endpoint: {q(endpoint)}
    access-key-id: {q(access_key)}
    access-key-secret: {q(secret_key)}
    bucket-name: {q(bucket)}
    use-ssl: {str(use_ssl).lower()}
    base-path: {q(base_path)}
    bucket-url: {q(f"{scheme}://{endpoint}/{bucket}")}
'''

redis_block = f'''redis:
    useCluster: false
    addr: {q(redis_addr)}
    password: {q(redis_password)}
    db: {redis_db}
    clusterAddrs:
        - {q(redis_addr)}
'''

redis_list_block = f'''redis-list:
    - name: cache
      useCluster: false
      addr: {q(redis_addr)}
      password: {q(redis_password)}
      db: {redis_db}
      clusterAddrs:
          - {q(redis_addr)}
'''

for filename in ("config.init.yaml", "config.yaml"):
    path = Path(filename)
    if not path.exists():
        continue
    text = path.read_text(encoding="utf-8")
    text, count = re.subn(r"(?m)^minio:\n(?:^[ \t][^\n]*\n?)*", block, text, count=1)
    if count != 1:
        raise SystemExit(f"{filename}: 未找到 minio 配置块")
    text, count = re.subn(r"(?ms)^redis:\n.*?(?=^redis-list:)", redis_block, text, count=1)
    if count != 1:
        raise SystemExit(f"{filename}: 未找到 redis 配置块")
    text, count = re.subn(r"(?ms)^redis-list:\n.*?(?=^mongo:)", redis_list_block, text, count=1)
    if count != 1:
        raise SystemExit(f"{filename}: 未找到 redis-list 配置块")
    text, count = re.subn(r"(?m)^(\s*use-redis:\s*).*$", rf"\g<1>{str(use_redis).lower()}", text, count=1)
    if count != 1:
        raise SystemExit(f"{filename}: 未找到 system.use-redis 配置")
    text, count = re.subn(r"(?m)^(\s*oss-type:\s*).*$", r"\g<1>minio", text, count=1)
    if count != 1:
        raise SystemExit(f"{filename}: 未找到 system.oss-type 配置")
    path.write_text(text, encoding="utf-8")
    path.chmod(0o600)
PY

echo "Redis 与 RustFS/MinIO 配置已写入 config.init.yaml 和 config.yaml"
