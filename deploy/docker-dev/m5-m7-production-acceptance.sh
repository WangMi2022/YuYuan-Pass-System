#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

ENV_FILE="${ENV_FILE:-.env}"
[ -f "$ENV_FILE" ] || { printf '[ERROR] missing %s\n' "$ENV_FILE" >&2; exit 2; }

for command_name in python3 psql; do
  command -v "$command_name" >/dev/null 2>&1 || {
    printf '[ERROR] missing command: %s\n' "$command_name" >&2
    exit 2
  }
done

set -a
# shellcheck disable=SC1090
. "$ENV_FILE"
set +a

exec python3 tests/m5-m7-production-acceptance.py "$@"
