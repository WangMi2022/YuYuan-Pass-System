#!/usr/bin/env python3
"""Production business acceptance for smart milestones M5, M6, and M7.

The script is intentionally opt-in because it creates isolated test records,
temporarily extends role 8881, exercises scheduled report delivery, and then
restores permissions and removes every test record.
"""

import argparse
import base64
import json
import os
import secrets
import signal
import subprocess
import sys
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional, Tuple


ADMIN_AUTHORITY = 888
LIMITED_AUTHORITY = 8881
EXPECTED_TOOLS = {
    "asset.search",
    "asset.detail",
    "asset.risk.list",
    "asset.warranty.expiring",
    "asset.custodian.summary",
    "asset.operation.summary",
    "invoice.summary",
    "invoice.pending_reviews",
    "invoice.failed_recognitions",
    "invoice.provider_quality",
    "schedule.today",
    "announcement.unread",
}
LIMITED_RULES = [
    ("/user/setUserAuthority", "POST"),
    ("/smart/copilot/sessions", "GET"),
    ("/smart/copilot/session", "GET"),
    ("/smart/copilot/session", "DELETE"),
    ("/smart/copilot/tools", "GET"),
    ("/smart/copilot/query", "POST"),
    ("/smart/copilot/queryStream", "POST"),
    ("/smart/drafts", "GET"),
    ("/smart/operation/assets", "GET"),
    ("/smart/announcement/extract", "POST"),
    ("/smart/operation/draft", "POST"),
    ("/smart/draft/accept", "POST"),
    ("/asset/list", "GET"),
]


class AcceptanceError(RuntimeError):
    pass


def stop_for_signal(signum: int, _frame: Any) -> None:
    raise KeyboardInterrupt(f"received signal {signum}")


def passed(name: str, detail: str = "") -> None:
    suffix = f" - {detail}" if detail else ""
    print(f"[PASS] {name}{suffix}", flush=True)


def require(condition: bool, name: str, detail: str = "") -> None:
    if not condition:
        suffix = f": {detail}" if detail else ""
        raise AcceptanceError(f"{name}{suffix}")
    passed(name, detail)


class HttpResult:
    def __init__(self, status: int, headers: Any, body: bytes, payload: Any) -> None:
        self.status = status
        self.headers = headers
        self.body = body
        self.payload = payload


class ApiClient:
    def __init__(self, base_url: str, token: str = "") -> None:
        self.base_url = base_url.rstrip("/")
        self.token = token

    def request(
        self,
        method: str,
        path: str,
        data: Optional[Any] = None,
        timeout: int = 75,
    ) -> HttpResult:
        body = None if data is None else json.dumps(data, ensure_ascii=False).encode("utf-8")
        headers = {"Accept": "application/json", "User-Agent": "m5-m7-production-acceptance"}
        if body is not None:
            headers["Content-Type"] = "application/json"
        if self.token:
            headers["x-token"] = self.token
        request = urllib.request.Request(
            self.base_url + path,
            data=body,
            headers=headers,
            method=method,
        )
        try:
            response = urllib.request.urlopen(request, timeout=timeout)
        except urllib.error.HTTPError as error:
            raw = error.read()
            response_headers = error.headers
            status = error.code
        else:
            raw = response.read()
            response_headers = response.headers
            status = response.status
        try:
            payload = json.loads(raw.decode("utf-8")) if raw else None
        except (UnicodeDecodeError, json.JSONDecodeError):
            payload = None
        return HttpResult(status=status, headers=response_headers, body=raw, payload=payload)

    def ok(self, method: str, path: str, data: Optional[Any] = None, timeout: int = 75) -> Any:
        result = self.request(method, path, data=data, timeout=timeout)
        message = result.payload.get("msg", "invalid response") if isinstance(result.payload, dict) else "invalid response"
        if result.status != 200 or not isinstance(result.payload, dict) or result.payload.get("code") != 0:
            raise AcceptanceError(f"{method} {path} failed: http={result.status} msg={message}")
        return result.payload.get("data")

    def fails(self, method: str, path: str, data: Optional[Any] = None) -> bool:
        result = self.request(method, path, data=data)
        return result.status != 200 or not isinstance(result.payload, dict) or result.payload.get("code") != 0


class AcceptanceRun:
    def __init__(self, base_url: str) -> None:
        self.base_url = base_url.rstrip("/")
        self.run_tag = f"m5m7_acceptance_{datetime.now().strftime('%Y%m%d%H%M%S')}_{secrets.token_hex(3)}"
        self.admin_username = os.environ.get("GVA_ADMIN_USERNAME", "admin")
        self.admin_password = os.environ.get("GVA_ADMIN_PASSWORD", "")
        self.admin_client = None  # type: Optional[ApiClient]
        self.user_client = None  # type: Optional[ApiClient]
        self.original_limited_rules = None  # type: Optional[List[Dict[str, str]]]
        self.temp_user_id = 0
        self.temp_username = ""
        self.temp_password = ""
        self.announcement_id = 0
        self.asset_id = 0
        self.operation_order_id = 0

    def login(self, username: str, password: str) -> Tuple[ApiClient, Dict[str, Any]]:
        client = ApiClient(self.base_url)
        data = client.ok(
            "POST",
            "/base/login",
            {"username": username, "password": password, "captcha": "", "captchaId": ""},
        )
        require(isinstance(data, dict) and bool(data.get("token")), "login token issued")
        client.token = data["token"]
        return client, data.get("user") or {}

    def db(self, sql: str, variables: Optional[Dict[str, Any]] = None) -> str:
        required = ["GVA_PG_HOST", "GVA_PG_PORT", "GVA_PG_USER", "GVA_PG_DB", "GVA_PG_PASSWORD"]
        missing = [name for name in required if not os.environ.get(name)]
        if missing:
            raise AcceptanceError("missing database environment variables: " + ",".join(missing))
        command = [
            "psql",
            "-X",
            "-qAt",
            "-v",
            "ON_ERROR_STOP=1",
            "-h",
            os.environ["GVA_PG_HOST"],
            "-p",
            os.environ["GVA_PG_PORT"],
            "-U",
            os.environ["GVA_PG_USER"],
            "-d",
            os.environ["GVA_PG_DB"],
        ]
        for key, value in (variables or {}).items():
            command.extend(["-v", f"{key}={value}"])
        environment = os.environ.copy()
        environment["PGPASSWORD"] = os.environ["GVA_PG_PASSWORD"]
        result = subprocess.run(
            command,
            input=sql,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            universal_newlines=True,
            env=environment,
            check=False,
        )
        if result.returncode != 0:
            message = result.stderr.strip().splitlines()[-1] if result.stderr.strip() else "psql failed"
            raise AcceptanceError(f"database command failed: {message}")
        return result.stdout.strip()

    def update_limited_policy(self, rules: List[Dict[str, str]]) -> None:
        if self.admin_client is None:
            raise AcceptanceError("admin client is unavailable")
        self.admin_client.ok(
            "POST",
            "/casbin/updateCasbin",
            {"authorityId": LIMITED_AUTHORITY, "casbinInfos": rules},
        )

    def switch_authority(self, authority_id: int) -> None:
        if self.user_client is None:
            raise AcceptanceError("temporary user client is unavailable")
        result = self.user_client.request("POST", "/user/setUserAuthority", {"authorityId": authority_id})
        message = result.payload.get("msg", "invalid response") if isinstance(result.payload, dict) else "invalid response"
        if result.status != 200 or not isinstance(result.payload, dict) or result.payload.get("code") != 0:
            raise AcceptanceError(f"switch authority failed: http={result.status} msg={message}")
        token = result.headers.get("new-token")
        require(bool(token), f"authority switched to {authority_id}")
        self.user_client.token = token

    def prepare(self) -> None:
        require(bool(self.admin_password), "admin password environment is configured")
        self.admin_client, admin_user = self.login(self.admin_username, self.admin_password)
        require(int(admin_user.get("authorityId", 0)) == ADMIN_AUTHORITY, "admin authority is 888")

        menu = self.admin_client.ok("POST", "/menu/getMenu", {})
        require(menu is not None, "menu endpoint uses POST")

        policies = self.admin_client.ok(
            "POST", "/casbin/getPolicyPathByAuthorityId", {"authorityId": LIMITED_AUTHORITY}
        )
        self.original_limited_rules = list((policies or {}).get("paths") or [])
        merged = {(item["path"], item["method"]): item for item in self.original_limited_rules}
        for path, method in LIMITED_RULES:
            merged[(path, method)] = {"path": path, "method": method}
        self.update_limited_policy(list(merged.values()))
        passed("temporary limited-role policy installed")

        self.temp_username = self.run_tag
        self.temp_password = secrets.token_urlsafe(24)
        created = self.admin_client.ok(
            "POST",
            "/user/admin_register",
            {
                "userName": self.temp_username,
                "passWord": self.temp_password,
                "nickName": "M5-M7 Acceptance",
                "authorityId": ADMIN_AUTHORITY,
                "authorityIds": [ADMIN_AUTHORITY, LIMITED_AUTHORITY],
                "enable": 1,
                "email": "",
            },
        )
        user = (created or {}).get("user") or {}
        self.temp_user_id = int(user.get("ID", 0))
        require(self.temp_user_id > 0, "temporary acceptance user created")
        self.user_client, temp_user = self.login(self.temp_username, self.temp_password)
        require(int(temp_user.get("ID", 0)) == self.temp_user_id, "temporary acceptance user login")

    def accept_m5(self) -> None:
        assert self.user_client is not None
        tools = self.user_client.ok("GET", "/smart/copilot/tools") or []
        names = {item.get("name") for item in tools}
        require(names == EXPECTED_TOOLS, "M5 admin has all 12 read-only tools", f"count={len(names)}")

        asset_page = self.user_client.ok("GET", "/asset/list?page=1&pageSize=1") or {}
        assets = asset_page.get("list") or []
        require(bool(assets), "M5 asset data is available for citation acceptance")
        asset_id = int(assets[0].get("ID", 0))

        initial_sessions = self.user_client.ok("GET", "/smart/copilot/sessions") or []
        query = self.user_client.ok(
            "POST", "/smart/copilot/query", {"question": f"查看资产 ID {asset_id} 详情", "sessionId": 0}
        )
        session_id = int((query or {}).get("sessionId", 0))
        require(session_id > 0 and query.get("readOnly") is True, "M5 read-only query creates a session")
        citations = query.get("citations") or []
        require(len(citations) == 1 and int(citations[0].get("id", 0)) == asset_id, "M5 citation points to asset")

        detail = self.user_client.ok("GET", f"/smart/copilot/session?id={session_id}") or {}
        require(len(detail.get("messages") or []) == 2, "M5 session stores one atomic message pair")

        sse = self.user_client.request(
            "POST",
            "/smart/copilot/queryStream",
            {"question": "查询资产库存", "sessionId": session_id},
        )
        require(
            sse.status == 200
            and "text/event-stream" in (sse.headers.get("Content-Type") or "")
            and sse.body.startswith(b"data: "),
            "M5 SSE contract",
        )

        before_rejection = len(self.user_client.ok("GET", "/smart/copilot/sessions") or [])
        require(
            self.user_client.fails("POST", "/smart/copilot/query", {"question": "删除资产", "sessionId": 0}),
            "M5 write intent is rejected",
        )
        after_rejection = len(self.user_client.ok("GET", "/smart/copilot/sessions") or [])
        require(before_rejection == after_rejection, "M5 rejected write creates no empty session")

        self.user_client.ok("DELETE", f"/smart/copilot/session?id={session_id}")
        require(self.user_client.fails("GET", f"/smart/copilot/session?id={session_id}"), "M5 session deletion")
        require(len(initial_sessions) == 0, "M5 temporary user started without historical sessions")

        self.switch_authority(LIMITED_AUTHORITY)
        limited_tools = self.user_client.ok("GET", "/smart/copilot/tools") or []
        limited_names = {item.get("name") for item in limited_tools}
        require(
            limited_names == {"asset.search", "asset.warranty.expiring", "asset.custodian.summary"},
            "M5 limited role filters tools",
            f"count={len(limited_names)}",
        )
        limited_before = len(self.user_client.ok("GET", "/smart/copilot/sessions") or [])
        require(
            self.user_client.fails("POST", "/smart/copilot/query", {"question": "查询发票汇总", "sessionId": 0}),
            "M5 limited role cannot read invoice tool",
        )
        limited_after = len(self.user_client.ok("GET", "/smart/copilot/sessions") or [])
        require(limited_before == limited_after, "M5 unauthorized query creates no session")
        self.switch_authority(ADMIN_AUTHORITY)

    def accept_deterministic_fallback(self) -> None:
        assert self.admin_client is not None
        assert self.user_client is not None
        original = self.admin_client.ok("GET", "/ai/providers") or {}
        disabled = json.loads(json.dumps(original))
        disabled["enabled"] = False
        self.admin_client.ok("PUT", "/ai/providers", disabled)
        session_id = 0
        try:
            query = self.user_client.ok(
                "POST", "/smart/copilot/query", {"question": "查询资产库存", "sessionId": 0}
            ) or {}
            session_id = int(query.get("sessionId", 0))
            require(session_id > 0 and query.get("modelUsed") is False, "M5 deterministic fallback with AI disabled")
            report = self.user_client.ok("POST", "/smartReport/generate") or {}
            require(report.get("generatedBy") == "deterministic", "M6 deterministic fallback with AI disabled")
        finally:
            self.admin_client.ok("PUT", "/ai/providers", original)
            restored = self.admin_client.ok("GET", "/ai/providers") or {}
            require(restored == original, "AI provider configuration restored")
            if session_id:
                self.user_client.ok("DELETE", f"/smart/copilot/session?id={session_id}")

    def accept_m6(self) -> None:
        assert self.user_client is not None
        report = self.user_client.ok("POST", "/smartReport/generate") or {}
        report_id = int(report.get("ID", 0))
        metrics = report.get("metrics") or {}
        require(report_id > 0, "M6 manual report generated")
        require(
            set(metrics) >= {"assets", "risks", "invoices", "collaboration", "system"},
            "M6 report contains all metric groups",
        )
        expected_metrics = dict(
            line.split("=", 1)
            for line in self.db(
                """
                SELECT 'assets_created=' || COUNT(*) FROM assets
                  WHERE deleted_at IS NULL AND created_at >= CURRENT_DATE AND created_at < CURRENT_DATE + 1;
                SELECT 'risks_open=' || COUNT(*) FROM asset_risk_events
                  WHERE deleted_at IS NULL AND status IN ('open', 'acknowledged');
                SELECT 'invoices_pending=' || COUNT(*) FROM invoices
                  WHERE deleted_at IS NULL AND status = 'pending_review';
                """
            ).splitlines()
        )
        require(
            int(metrics["assets"]["created"]) == int(expected_metrics["assets_created"])
            and int(metrics["risks"]["open"]) == int(expected_metrics["risks_open"])
            and int(metrics["invoices"]["pendingReview"]) == int(expected_metrics["invoices_pending"]),
            "M6 report metrics match independent database counts",
        )
        report_list = self.user_client.ok("GET", "/smartReport/list?page=1&pageSize=20") or {}
        require(any(int(item.get("ID", 0)) == report_id for item in report_list.get("list") or []), "M6 report history")
        detail = self.user_client.ok("GET", f"/smartReport/detail?id={report_id}") or {}
        require(int(detail.get("ID", 0)) == report_id, "M6 report detail")

        now = datetime.now()
        delay_minutes = 1 if now.second <= 45 else 2
        delivery_time = (now + timedelta(minutes=delay_minutes)).strftime("%H:%M")
        subscription = self.user_client.ok(
            "PUT",
            "/smartReport/subscription",
            {"enabled": True, "deliveryTime": delivery_time, "channels": "in_app,email"},
        ) or {}
        require(subscription.get("channels") == "in_app,email", "M6 subscription upsert")
        deliveries = []  # type: List[Dict[str, Any]]
        channels = {}  # type: Dict[str, Dict[str, Any]]
        deadline = time.time() + 155
        while time.time() < deadline:
            deliveries = self.user_client.ok("GET", "/smartReport/deliveries") or []
            channels = {
                item.get("channel"): item
                for item in deliveries
                if int(item.get("reportId", 0)) == report_id
            }
            if (
                channels.get("in_app", {}).get("status") == "sent"
                and channels.get("email", {}).get("status") in {"sent", "failed"}
            ):
                break
            time.sleep(3)
        channels = {item.get("channel"): item for item in deliveries if int(item.get("reportId", 0)) == report_id}
        require(channels.get("in_app", {}).get("status") == "sent", "M6 scheduled in-app delivery")
        email = channels.get("email") or {}
        require(
            email.get("status") in {"sent", "failed"}
            and int(email.get("retryCount", 0)) >= 1
            and (email.get("status") == "sent" or bool(email.get("error"))),
            "M6 email delivery result is auditable",
            f"status={email.get('status')}",
        )

    def create_announcement(self) -> None:
        title = f"{self.run_tag} 资产盘点会议"
        content = "请于2026年08月20日下午14:30参加资产盘点会议。地点：会议室A。请提前提交参会名单。"
        title_base64 = base64.b64encode(title.encode("utf-8")).decode("ascii")
        content_base64 = base64.b64encode(content.encode("utf-8")).decode("ascii")
        output = self.db(
            """
            INSERT INTO gva_announcements_info
                (created_at, updated_at, title, content, user_id, status, published_at)
            VALUES (
                NOW(),
                NOW(),
                convert_from(decode(:'title_base64', 'base64'), 'UTF8'),
                convert_from(decode(:'content_base64', 'base64'), 'UTF8'),
                :user_id,
                'published',
                NOW()
            )
            RETURNING id;
            """,
            {
                "title_base64": title_base64,
                "content_base64": content_base64,
                "user_id": self.temp_user_id,
            },
        )
        self.announcement_id = int(output.splitlines()[-1])
        require(self.announcement_id > 0, "M7 test announcement created")

    def create_asset(self) -> None:
        assert self.user_client is not None
        category_id = int(self.db("SELECT id FROM asset_categories WHERE enabled = true ORDER BY id LIMIT 1;"))
        asset = self.user_client.ok(
            "POST",
            "/asset/create",
            {
                "assetCode": self.run_tag.upper(),
                "name": "M5-M7 Acceptance Asset",
                "categoryId": category_id,
                "quantity": 1,
                "unit": "item",
                "unitPrice": 100,
                "currentValue": 100,
                "remarks": self.run_tag,
            },
        ) or {}
        self.asset_id = int(asset.get("ID", 0))
        require(self.asset_id > 0, "M7 test asset created")

    def accept_m7(self) -> None:
        assert self.user_client is not None
        self.create_announcement()
        self.create_asset()

        extracted = self.user_client.ok(
            "POST", "/smart/announcement/extract", {"announcementId": self.announcement_id}
        ) or {}
        schedule_draft_id = int(extracted.get("ID", 0))
        payload = extracted.get("payload") or {}
        require(
            schedule_draft_id > 0
            and payload.get("date") == "2026-08-20"
            and payload.get("time") == "14:30"
            and payload.get("location") == "会议室A"
            and bool(payload.get("todos")),
            "M7 announcement schedule extraction",
        )

        self.db(
            "UPDATE smart_drafts SET expires_at = NOW() - INTERVAL '1 minute' WHERE id = :draft_id AND user_id = :user_id;",
            {"draft_id": schedule_draft_id, "user_id": self.temp_user_id},
        )
        refreshed = self.user_client.ok(
            "POST", "/smart/announcement/extract", {"announcementId": self.announcement_id}
        ) or {}
        require(
            int(refreshed.get("ID", 0)) == schedule_draft_id and refreshed.get("status") == "draft",
            "M7 expired announcement draft refreshes in place",
        )

        candidates = self.user_client.ok("GET", "/smart/operation/assets?operationType=inbound") or []
        require(any(int(item.get("ID", 0)) == self.asset_id for item in candidates), "M7 operation asset candidates")
        operation_draft = self.user_client.ok(
            "POST",
            "/smart/operation/draft",
            {
                "operationType": "inbound",
                "instruction": "Create inbound draft only",
                "assetIds": [self.asset_id],
                "targetLocation": "Acceptance Temporary Location",
                "remarks": self.run_tag,
            },
        ) or {}
        operation_draft_id = int(operation_draft.get("ID", 0))
        require(operation_draft_id > 0 and operation_draft.get("status") == "draft", "M7 operation draft generated")

        self.switch_authority(LIMITED_AUTHORITY)
        require(
            self.user_client.fails("GET", "/smart/operation/assets?operationType=inbound"),
            "M7 limited role candidate permission is rechecked",
        )
        require(
            self.user_client.fails("POST", "/smart/draft/accept", {"id": schedule_draft_id}),
            "M7 limited role schedule create permission is rechecked",
        )
        require(
            self.user_client.fails("POST", "/smart/draft/accept", {"id": operation_draft_id}),
            "M7 limited role operation create permission is rechecked",
        )
        require(
            self.user_client.fails(
                "POST",
                "/smart/operation/draft",
                {"operationType": "inbound", "assetIds": [self.asset_id], "targetLocation": "Denied"},
            ),
            "M7 limited role cannot generate operation drafts",
        )
        self.switch_authority(ADMIN_AUTHORITY)

        schedule = self.user_client.ok("POST", "/smart/draft/accept", {"id": schedule_draft_id}) or {}
        require(int(schedule.get("id", 0)) > 0, "M7 schedule draft accepted")

        def accept_operation() -> HttpResult:
            client = ApiClient(self.base_url, self.user_client.token)
            return client.request("POST", "/smart/draft/accept", {"id": operation_draft_id})

        with ThreadPoolExecutor(max_workers=2) as executor:
            results = [future.result() for future in [executor.submit(accept_operation), executor.submit(accept_operation)]]
        success = [item for item in results if item.status == 200 and isinstance(item.payload, dict) and item.payload.get("code") == 0]
        rejected = [item for item in results if item not in success]
        require(len(success) == 1 and len(rejected) == 1, "M7 concurrent confirmation allows exactly one winner")
        order = success[0].payload.get("data") or {}
        self.operation_order_id = int(order.get("ID", 0))
        require(
            self.operation_order_id > 0 and order.get("status") == "draft",
            "M7 accepted operation remains submit=false draft",
        )

    def cleanup_database(self) -> None:
        if not self.temp_user_id:
            return
        self.db(
            """
            DELETE FROM work_schedule_notifications WHERE user_id = :user_id;
            DELETE FROM work_schedules WHERE user_id = :user_id;
            DELETE FROM asset_operation_records WHERE operator_id = :user_id;
            DELETE FROM asset_operation_items
              WHERE order_id IN (SELECT id FROM asset_operation_orders WHERE created_by = :user_id);
            DELETE FROM asset_operation_orders WHERE created_by = :user_id;
            DELETE FROM asset_risk_events WHERE asset_id = :asset_id;
            DELETE FROM smart_report_deliveries WHERE user_id = :user_id;
            DELETE FROM smart_daily_reports WHERE user_id = :user_id;
            DELETE FROM smart_report_subscriptions WHERE user_id = :user_id;
            DELETE FROM smart_drafts WHERE user_id = :user_id;
            DELETE FROM ai_copilot_messages WHERE user_id = :user_id;
            DELETE FROM ai_copilot_sessions WHERE user_id = :user_id;
            DELETE FROM ai_model_invocations WHERE user_id = :user_id;
            DELETE FROM gva_announcement_reads
              WHERE user_id = :user_id OR announcement_id = :announcement_id;
            DELETE FROM gva_announcements_info WHERE id = :announcement_id;
            DELETE FROM assets WHERE id = :asset_id;
            DELETE FROM sys_operation_records WHERE user_id = :user_id;
            DELETE FROM sys_login_logs WHERE user_id = :user_id;
            DELETE FROM sys_user_authority WHERE sys_user_id = :user_id;
            DELETE FROM sys_users WHERE id = :user_id;
            """,
            {
                "user_id": self.temp_user_id,
                "announcement_id": self.announcement_id,
                "asset_id": self.asset_id,
            },
        )
        remaining = self.db(
            """
            SELECT
              (SELECT COUNT(*) FROM sys_users WHERE id = :user_id OR username = :'username') +
              (SELECT COUNT(*) FROM assets WHERE id = :asset_id) +
              (SELECT COUNT(*) FROM gva_announcements_info WHERE id = :announcement_id) +
              (SELECT COUNT(*) FROM asset_operation_orders WHERE created_by = :user_id) +
              (SELECT COUNT(*) FROM work_schedules WHERE user_id = :user_id) +
              (SELECT COUNT(*) FROM smart_report_deliveries WHERE user_id = :user_id) +
              (SELECT COUNT(*) FROM smart_daily_reports WHERE user_id = :user_id) +
              (SELECT COUNT(*) FROM smart_report_subscriptions WHERE user_id = :user_id) +
              (SELECT COUNT(*) FROM smart_drafts WHERE user_id = :user_id) +
              (SELECT COUNT(*) FROM ai_copilot_messages WHERE user_id = :user_id) +
              (SELECT COUNT(*) FROM ai_copilot_sessions WHERE user_id = :user_id);
            """,
            {
                "user_id": self.temp_user_id,
                "username": self.temp_username,
                "announcement_id": self.announcement_id,
                "asset_id": self.asset_id,
            },
        )
        require(remaining == "0", "all acceptance test data removed")

    def cleanup(self) -> None:
        errors = []  # type: List[str]
        if self.original_limited_rules is not None:
            try:
                if self.admin_client is None:
                    self.admin_client, _ = self.login(self.admin_username, self.admin_password)
                self.update_limited_policy(self.original_limited_rules)
                restored = self.admin_client.ok(
                    "POST", "/casbin/getPolicyPathByAuthorityId", {"authorityId": LIMITED_AUTHORITY}
                )
                restored_rules = {
                    (item["path"], item["method"])
                    for item in ((restored or {}).get("paths") or [])
                }
                original_rules = {
                    (item["path"], item["method"])
                    for item in self.original_limited_rules
                }
                require(restored_rules == original_rules, "limited-role policy restored")
            except Exception as error:  # cleanup must continue
                errors.append(f"policy restore: {error}")
        try:
            self.cleanup_database()
        except Exception as error:  # cleanup must report every failure
            errors.append(f"database cleanup: {error}")
        if errors:
            raise AcceptanceError("; ".join(errors))

    def run(self) -> None:
        primary_error = None  # type: Optional[BaseException]
        try:
            self.prepare()
            self.accept_m5()
            self.accept_deterministic_fallback()
            self.accept_m6()
            self.accept_m7()
        except BaseException as error:
            primary_error = error
        try:
            self.cleanup()
        except Exception as cleanup_error:
            if primary_error is None:
                primary_error = cleanup_error
            else:
                primary_error = AcceptanceError(f"{primary_error}; cleanup: {cleanup_error}")
        if primary_error is not None:
            raise primary_error
        print("M5-M7 production business acceptance passed.", flush=True)


def main() -> int:
    parser = argparse.ArgumentParser(description="Run mutating M5-M7 production business acceptance")
    parser.add_argument("--execute", action="store_true", help="confirm temporary production test mutations")
    parser.add_argument("--base-url", default=os.environ.get("SMART_ACCEPTANCE_BASE_URL", "http://127.0.0.1:8888"))
    arguments = parser.parse_args()
    if not arguments.execute:
        print("Refusing to mutate production without --execute.", file=sys.stderr)
        return 2
    signal.signal(signal.SIGTERM, stop_for_signal)
    if hasattr(signal, "SIGHUP"):
        signal.signal(signal.SIGHUP, stop_for_signal)
    try:
        AcceptanceRun(base_url=arguments.base_url).run()
    except Exception as error:
        print(f"[FAIL] {error}", file=sys.stderr, flush=True)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
