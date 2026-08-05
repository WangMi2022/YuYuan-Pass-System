#!/usr/bin/env bash
set -euo pipefail

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
DEPLOY_DIR=$(cd "${TEST_DIR}/.." && pwd)
SCRIPT_UNDER_TEST="${DEPLOY_DIR}/release-acceptance.sh"
MOCK_BIN="${TEST_DIR}/mocks"
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

printf 'WEB_PORT=8080\nSERVER_PORT=8888\nGVA_TEST_SECRET=must-not-be-exported\n' > "${TEMP_DIR}/.env"
: > "${TEMP_DIR}/docker-compose.yml"

passed=0
failed=0

verify_mocks() {
  local docker_output

  if ! docker_output=$(MOCK_SCENARIO=success "${MOCK_BIN}/docker" compose --env-file "${TEMP_DIR}/.env" -f "${TEMP_DIR}/docker-compose.yml" config --services) \
    || ! printf '%s\n' "$docker_output" | grep -Fxq 'web'; then
    printf '[FAIL] 测试基础设施：Docker mock 无法执行\n' >&2
    exit 1
  fi

  if ! MOCK_SCENARIO=success "${MOCK_BIN}/curl" -fsS --max-time 1 \
    -D "${TEMP_DIR}/mock.headers" -o "${TEMP_DIR}/mock.body" -w '%{http_code}' \
    'http://server.test/health' >/dev/null; then
    printf '[FAIL] 测试基础设施：curl mock 无法执行\n' >&2
    exit 1
  fi
}

verify_mocks

run_case() {
  local name="$1"
  local scenario="$2"
  local expected_status="$3"
  local expected_text="$4"
  local output
  local status

  set +e
  output=$(
    PATH="${MOCK_BIN}:$PATH" \
      MOCK_SCENARIO="$scenario" \
      ENV_FILE="${TEMP_DIR}/.env" \
      COMPOSE_FILE="${TEMP_DIR}/docker-compose.yml" \
      WEB_URL="http://web.test" \
      SERVER_URL="http://server.test" \
      MAX_ATTEMPTS=1 \
      STARTUP_MAX_ATTEMPTS="${CASE_STARTUP_MAX_ATTEMPTS:-1}" \
      RETRY_INTERVAL_SECONDS=0 \
      MOCK_STATE_DIR="$TEMP_DIR" \
      bash "$SCRIPT_UNDER_TEST" 2>&1
  )
  status=$?
  set -e

  if [ "$status" -eq "$expected_status" ] \
    && printf '%s' "$output" | grep -Fq "$expected_text" \
    && printf '%s' "$output" | grep -Fq '[PASS] compose.services'; then
    printf '[PASS] %s\n' "$name"
    passed=$((passed + 1))
    return
  fi

  printf '[FAIL] %s\n' "$name" >&2
  printf '  expected status=%s and text=%s\n' "$expected_status" "$expected_text" >&2
  printf '  actual status=%s\n%s\n' "$status" "$output" >&2
  failed=$((failed + 1))
}

run_config_case() {
  local name="$1"
  local variable_name="$2"
  local variable_value="$3"
  local expected_text="$4"
  local output
  local status

  set +e
  output=$(
    env \
      PATH="${MOCK_BIN}:$PATH" \
      ENV_FILE="${TEMP_DIR}/.env" \
      COMPOSE_FILE="${TEMP_DIR}/docker-compose.yml" \
      WEB_URL="http://web.test" \
      SERVER_URL="http://server.test" \
      MAX_ATTEMPTS=1 \
      STARTUP_MAX_ATTEMPTS=1 \
      RETRY_INTERVAL_SECONDS=0 \
      "$variable_name=$variable_value" \
      bash "$SCRIPT_UNDER_TEST" 2>&1
  )
  status=$?
  set -e

  if [ "$status" -eq 2 ] && printf '%s' "$output" | grep -Fq "$expected_text"; then
    printf '[PASS] %s\n' "$name"
    passed=$((passed + 1))
    return
  fi

  printf '[FAIL] %s\n' "$name" >&2
  printf '  expected status=2 and text=%s\n' "$expected_text" >&2
  printf '  actual status=%s\n%s\n' "$status" "$output" >&2
  failed=$((failed + 1))
}

run_case "全部检查通过" "success" 0 "上线验收通过"
CASE_STARTUP_MAX_ATTEMPTS=2 run_case "后端慢启动时在独立窗口内重试成功" "server-health-delayed" 0 "上线验收通过"
run_case "缺少运行中的服务时失败" "missing-service" 1 "[FAIL] containers.running"
run_case "容器出现异常重启时失败" "restart-loop" 1 "[FAIL] containers.restarts"
run_case "后端健康接口不可用时失败" "server-health-fail" 1 "[FAIL] api.direct"
run_case "数据库未初始化时失败" "db-not-ready" 1 "[FAIL] api.database"
run_case "Web 返回错误首页时失败" "invalid-web" 1 "[FAIL] web.home"
run_case "Web API 反代断链时失败" "proxy-fail" 1 "[FAIL] api.proxy"
run_case "前端静态资源不可用时失败" "asset-fail" 1 "[FAIL] web.asset"
run_case "前端静态资源 MIME 错误时失败" "asset-wrong-mime" 1 "[FAIL] web.asset"
run_config_case "非法请求超时时间返回配置错误" "CURL_MAX_TIME" "abc" "CURL_MAX_TIME 必须是正整数"
run_config_case "非法启动等待次数返回配置错误" "STARTUP_MAX_ATTEMPTS" "0" "STARTUP_MAX_ATTEMPTS 必须大于 0"
run_config_case "非法重试间隔返回配置错误" "RETRY_INTERVAL_SECONDS" "-1" "RETRY_INTERVAL_SECONDS 必须是非负整数"
run_config_case "URL 中的账号凭据被拒绝" "WEB_URL" "https://user:secret@web.test" "WEB_URL 不能为空或包含账号凭据"

printf '\n测试汇总: passed=%d failed=%d\n' "$passed" "$failed"
[ "$failed" -eq 0 ]
