#!/usr/bin/env bash
set -uo pipefail

cd "$(dirname "$0")"

ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
EXPECTED_SERVICES="${EXPECTED_SERVICES:-server web}"
MAX_ATTEMPTS="${MAX_ATTEMPTS:-15}"
STARTUP_MAX_ATTEMPTS="${STARTUP_MAX_ATTEMPTS:-45}"
RETRY_INTERVAL_SECONDS="${RETRY_INTERVAL_SECONDS:-2}"
CURL_MAX_TIME="${CURL_MAX_TIME:-5}"
MAX_RESTART_COUNT="${MAX_RESTART_COUNT:-0}"
QUIET=0
CLI_WEB_URL=""
CLI_SERVER_URL=""

usage() {
  cat <<'EOF'
用法：./release-acceptance.sh [选项]

代码更新并重建容器后执行的只读上线验收。所有检查通过时退出码为 0。

选项：
  --web-url URL       覆盖 Web 地址
  --server-url URL    覆盖 Server 直连地址
  --quiet             只输出失败项和最终结果
  -h, --help          显示帮助

可配置环境变量：
  ENV_FILE COMPOSE_FILE EXPECTED_SERVICES MAX_ATTEMPTS
  STARTUP_MAX_ATTEMPTS RETRY_INTERVAL_SECONDS CURL_MAX_TIME
  MAX_RESTART_COUNT
EOF
}

configuration_error() {
  printf '[ERROR] %s\n' "$1" >&2
  exit 2
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --web-url)
      [ "$#" -ge 2 ] || configuration_error "--web-url 缺少 URL"
      CLI_WEB_URL="$2"
      shift 2
      ;;
    --server-url)
      [ "$#" -ge 2 ] || configuration_error "--server-url 缺少 URL"
      CLI_SERVER_URL="$2"
      shift 2
      ;;
    --quiet)
      QUIET=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      configuration_error "未知参数：$1"
      ;;
  esac
done

[ -f "$ENV_FILE" ] || configuration_error "缺少 ${ENV_FILE}，无法读取 Compose 运行配置"
[ -f "$COMPOSE_FILE" ] || configuration_error "缺少 ${COMPOSE_FILE}"

read_env_port() {
  local setting_name="$1"
  (
    set +u
    # 在短生命周期子进程中读取端口，避免将 .env 密钥导出给验收子进程。
    # shellcheck disable=SC1090
    . "$ENV_FILE" >/dev/null 2>&1 || exit 1
    case "$setting_name" in
      WEB_PORT) printf '%s' "${WEB_PORT:-}" ;;
      SERVER_PORT) printf '%s' "${SERVER_PORT:-}" ;;
      *) exit 2 ;;
    esac
  )
}

if ! ENV_WEB_PORT=$(read_env_port WEB_PORT) || ! ENV_SERVER_PORT=$(read_env_port SERVER_PORT); then
  configuration_error "${ENV_FILE} 不是有效的环境变量文件"
fi

WEB_URL="${CLI_WEB_URL:-${WEB_URL:-http://127.0.0.1:${ENV_WEB_PORT:-8080}}}"
SERVER_URL="${CLI_SERVER_URL:-${SERVER_URL:-http://127.0.0.1:${ENV_SERVER_PORT:-8888}}}"
WEB_URL="${WEB_URL%/}"
SERVER_URL="${SERVER_URL%/}"

for command_name in docker curl grep sed tr mktemp head sleep; do
  command -v "$command_name" >/dev/null 2>&1 || configuration_error "缺少命令：${command_name}"
done

validate_base_url() {
  local name="$1"
  local value="$2"
  local authority

  case "$value" in
    http://*|https://*) ;;
    *) configuration_error "${name} 必须使用 http:// 或 https://" ;;
  esac

  authority="${value#*://}"
  authority="${authority%%/*}"
  case "$authority" in
    ''|*@*) configuration_error "${name} 不能为空或包含账号凭据" ;;
  esac
  case "$value" in
    *\?*|*\#*|*" "*|*$'\t'*|*$'\r'*|*$'\n'*)
      configuration_error "${name} 不能包含空白、查询参数或片段"
      ;;
  esac
}

validate_base_url "WEB_URL" "$WEB_URL"
validate_base_url "SERVER_URL" "$SERVER_URL"

case "$MAX_ATTEMPTS" in
  ''|*[!0-9]*) configuration_error "MAX_ATTEMPTS 必须是正整数" ;;
esac
[ "$MAX_ATTEMPTS" -gt 0 ] || configuration_error "MAX_ATTEMPTS 必须大于 0"

case "$STARTUP_MAX_ATTEMPTS" in
  ''|*[!0-9]*) configuration_error "STARTUP_MAX_ATTEMPTS 必须是正整数" ;;
esac
[ "$STARTUP_MAX_ATTEMPTS" -gt 0 ] || configuration_error "STARTUP_MAX_ATTEMPTS 必须大于 0"

case "$MAX_RESTART_COUNT" in
  ''|*[!0-9]*) configuration_error "MAX_RESTART_COUNT 必须是非负整数" ;;
esac

case "$RETRY_INTERVAL_SECONDS" in
  ''|*[!0-9]*) configuration_error "RETRY_INTERVAL_SECONDS 必须是非负整数" ;;
esac

case "$CURL_MAX_TIME" in
  ''|*[!0-9]*) configuration_error "CURL_MAX_TIME 必须是正整数" ;;
esac
[ "$CURL_MAX_TIME" -gt 0 ] || configuration_error "CURL_MAX_TIME 必须大于 0"

WORK_DIR=$(mktemp -d)
trap 'rm -rf "$WORK_DIR"' EXIT

passed=0
failed=0

pass_check() {
  passed=$((passed + 1))
  if [ "$QUIET" -eq 0 ]; then
    printf '[PASS] %-22s %s\n' "$1" "$2"
  fi
}

fail_check() {
  failed=$((failed + 1))
  printf '[FAIL] %-22s %s\n' "$1" "$2" >&2
}

fetch_url() {
  local method="$1"
  local url="$2"
  local body_file="$3"
  local header_file="$4"
  local error_file="$5"
  local attempt
  local http_status
  local max_attempts="${6:-$MAX_ATTEMPTS}"
  local -a request_args

  request_args=(-fsS --max-time "$CURL_MAX_TIME" -D "$header_file" -o "$body_file" -w '%{http_code}')
  if [ "$method" = "POST" ]; then
    request_args+=(-X POST)
  fi

  attempt=1
  while [ "$attempt" -le "$max_attempts" ]; do
    : > "$body_file"
    : > "$header_file"
    : > "$error_file"
    if http_status=$(curl "${request_args[@]}" "$url" 2>"$error_file") && [ "$http_status" = "200" ]; then
      return 0
    fi

    if [ "$attempt" -lt "$max_attempts" ] && [ "$RETRY_INTERVAL_SECONDS" != "0" ]; then
      sleep "$RETRY_INTERVAL_SECONDS"
    fi
    attempt=$((attempt + 1))
  done

  return 1
}

is_ok_response() {
  [ "$(tr -d '[:space:]' < "$1")" = '"ok"' ]
}

printf '发布后自动验收\n'
printf 'Web: %s\nServer: %s\n\n' "$WEB_URL" "$SERVER_URL"

declared_services=""
if declared_services=$(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" config --services 2>/dev/null); then
  missing_declared=""
  for service in $EXPECTED_SERVICES; do
    if ! printf '%s\n' "$declared_services" | grep -Fxq "$service"; then
      missing_declared="${missing_declared} ${service}"
    fi
  done
  if [ -z "$missing_declared" ]; then
    pass_check "compose.services" "期望服务均已声明：${EXPECTED_SERVICES}"
  else
    fail_check "compose.services" "Compose 缺少服务：${missing_declared# }"
  fi
else
  fail_check "compose.services" "无法读取 Compose 服务配置"
fi

running_services=""
if running_services=$(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps --services --filter status=running 2>/dev/null); then
  missing_running=""
  for service in $EXPECTED_SERVICES; do
    if ! printf '%s\n' "$running_services" | grep -Fxq "$service"; then
      missing_running="${missing_running} ${service}"
    fi
  done
  if [ -z "$missing_running" ]; then
    pass_check "containers.running" "期望容器均处于 running 状态"
  else
    fail_check "containers.running" "未运行的服务：${missing_running# }"
  fi
else
  fail_check "containers.running" "无法查询容器运行状态"
fi

restart_problems=""
for service in $EXPECTED_SERVICES; do
  container_id=$(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps -q "$service" 2>/dev/null || true)
  if [ -z "$container_id" ]; then
    restart_problems="${restart_problems} ${service}=unavailable"
    continue
  fi

  restart_count=$(docker inspect --format '{{.RestartCount}}' "$container_id" 2>/dev/null || true)
  case "$restart_count" in
    ''|*[!0-9]*)
      restart_problems="${restart_problems} ${service}=unknown"
      ;;
    *)
      if [ "$restart_count" -gt "$MAX_RESTART_COUNT" ]; then
        restart_problems="${restart_problems} ${service}=${restart_count}"
      fi
      ;;
  esac
done
if [ -z "$restart_problems" ]; then
  pass_check "containers.restarts" "容器重启次数未超过 ${MAX_RESTART_COUNT}"
else
  fail_check "containers.restarts" "异常重启或状态不可查：${restart_problems# }"
fi

api_body="${WORK_DIR}/api-health.body"
api_headers="${WORK_DIR}/api-health.headers"
api_error="${WORK_DIR}/api-health.error"
if fetch_url GET "${SERVER_URL}/health" "$api_body" "$api_headers" "$api_error" "$STARTUP_MAX_ATTEMPTS" && is_ok_response "$api_body"; then
  pass_check "api.direct" "Server /health 正常"
else
  fail_check "api.direct" "Server /health 未返回 HTTP 200 和 \"ok\""
fi

db_body="${WORK_DIR}/database.body"
db_headers="${WORK_DIR}/database.headers"
db_error="${WORK_DIR}/database.error"
if fetch_url POST "${SERVER_URL}/init/checkdb" "$db_body" "$db_headers" "$db_error" \
  && grep -Eq '"code"[[:space:]]*:[[:space:]]*0' "$db_body" \
  && grep -Eq '"needInit"[[:space:]]*:[[:space:]]*false' "$db_body"; then
  pass_check "api.database" "API 已完成数据库初始化"
else
  fail_check "api.database" "API 不可用或数据库尚未完成初始化"
fi

web_body="${WORK_DIR}/web-home.body"
web_headers="${WORK_DIR}/web-home.headers"
web_error="${WORK_DIR}/web-home.error"
web_home_valid=0
if fetch_url GET "${WEB_URL}/" "$web_body" "$web_headers" "$web_error" \
  && grep -Eqi '^Content-Type:[[:space:]]*text/html([;[:space:]]|$)' "$web_headers" \
  && grep -Fq '<title>资产管理中心</title>' "$web_body" \
  && grep -Fq 'id="app"' "$web_body"; then
  web_home_valid=1
  pass_check "web.home" "Web 首页内容与应用挂载点正常"
else
  fail_check "web.home" "Web 首页不可用或返回了错误页面"
fi

proxy_body="${WORK_DIR}/proxy-health.body"
proxy_headers="${WORK_DIR}/proxy-health.headers"
proxy_error="${WORK_DIR}/proxy-health.error"
if fetch_url GET "${WEB_URL}/api/health" "$proxy_body" "$proxy_headers" "$proxy_error" && is_ok_response "$proxy_body"; then
  pass_check "api.proxy" "Web → Nginx → Server 反代链路正常"
else
  fail_check "api.proxy" "Web /api/health 反代链路不可用"
fi

asset_path=""
if [ "$web_home_valid" -eq 1 ]; then
  asset_path=$(grep -Eo 'src="[^"]*assets/[^"]+\.js[^"]*"' "$web_body" | head -n 1 | sed -E 's/^src="([^"]+)"$/\1/' || true)
fi

if [ -z "$asset_path" ]; then
  fail_check "web.asset" "首页没有可验证的 Vite JavaScript 入口"
else
  case "$asset_path" in
    http://*|https://*) asset_url="$asset_path" ;;
    /*) asset_url="${WEB_URL}${asset_path}" ;;
    ./*) asset_url="${WEB_URL}/${asset_path#./}" ;;
    *) asset_url="${WEB_URL}/${asset_path}" ;;
  esac

  asset_body="${WORK_DIR}/web-asset.body"
  asset_headers="${WORK_DIR}/web-asset.headers"
  asset_error="${WORK_DIR}/web-asset.error"
  if fetch_url GET "$asset_url" "$asset_body" "$asset_headers" "$asset_error" \
    && [ -s "$asset_body" ] \
    && grep -Eqi '^Content-Type:[[:space:]]*((application|text)/(javascript|x-javascript))([;[:space:]]|$)' "$asset_headers"; then
    pass_check "web.asset" "当前版本 JavaScript 入口可访问且非空"
  else
    fail_check "web.asset" "当前版本 JavaScript 入口不可访问、为空或类型错误"
  fi
fi

printf '\n验收汇总: passed=%d failed=%d\n' "$passed" "$failed"
if [ "$failed" -eq 0 ]; then
  printf '上线验收通过：当前版本可以标记为上线成功。\n'
  exit 0
fi

printf '上线验收失败：禁止写入上线成功版本标记。\n' >&2
exit 1
