"""System API tests for failed-login and password-guessing detections."""

import re
import time
import uuid
from datetime import UTC, datetime

import pytest
from api.file_service.conftest import ACCESS_COOKIE
from constants import (
    SECURITY_LOGGING_MANAGER_CLI_BINARY,
    SECURITY_LOGGING_MANAGER_LOGIN,
)
from fixtures.logging_manager_cli import as_event_items, execute_logging_manager_json
from hyperiontf import CLIClient, RESTClient, expect

pytestmark = [
    pytest.mark.api,
    pytest.mark.Auth,
    pytest.mark.WebGateway,
    pytest.mark.LogAnalysis,
    pytest.mark.SecurityLoggingManager,
]

LOGIN_NAME_PATTERN = re.compile(r"sec-test-[0-9a-f]{12}")
SYSTEM_LOGIN_PATTERN = re.compile(r"[a-z_][a-z0-9_-]{0,31}")
AUTH_LOG_COMMAND = "--cmd getAlerts --page 1 --pageSize 25"
ACTIVE_UNACKNOWLEDGED_COMMAND = "--cmd getActiveUnacknowledgedEvents --page 1 --pageSize 25"
INVALID_TEST_CREDENTIAL = "invalid-system-test-password"
HTTP_OK = 200
SECURITY_EVENT_WAIT_SECONDS = 20.0


def new_test_login() -> str:
    """Create a bounded, unique username that cannot collide with a real account."""
    login = f"sec-test-{uuid.uuid4().hex[:12]}"
    if not LOGIN_NAME_PATTERN.fullmatch(login):
        raise ValueError
    return login


def submit_failed_login(
    api_client: RESTClient,
    login: str,
    password: str = INVALID_TEST_CREDENTIAL,
) -> None:
    """Submit one rejected credential pair through the installed login API."""
    is_unique_test_login = LOGIN_NAME_PATTERN.fullmatch(login) is not None
    is_known_test_login = (
        login == SECURITY_LOGGING_MANAGER_LOGIN
        and SYSTEM_LOGIN_PATTERN.fullmatch(login) is not None
    )
    if not is_unique_test_login and not is_known_test_login:
        raise ValueError
    response = api_client.post(
        path="/api/auth/login",
        payload={"login": login, "password": password},
    )
    expect(response.status).to_be(401)


def find_security_event(
    security_cli_client: CLIClient,
    category: str,
    *,
    created_after: datetime | None = None,
    status: str | None = None,
    command: str = AUTH_LOG_COMMAND,
) -> dict[str, object] | None:
    """Poll the manager CLI for a recent Log Analysis event in one category."""
    deadline = time.monotonic() + SECURITY_EVENT_WAIT_SECONDS
    while time.monotonic() < deadline:
        payload = execute_logging_manager_json(
            security_cli_client,
            SECURITY_LOGGING_MANAGER_CLI_BINARY,
            command,
        )
        for event in as_event_items(payload):
            is_recent = True
            if created_after is not None:
                event_time_value = event.get("LastTimestamp") or event.get("CreatedAt")
                if not isinstance(event_time_value, str):
                    continue
                try:
                    event_time = datetime.fromisoformat(event_time_value.replace("Z", "+00:00"))
                except ValueError:
                    continue
                is_recent = event_time >= created_after
            if (
                event.get("Source") == "log-analysis"
                and event.get("Category") == category
                and is_recent
                and (status is None or event.get("Status") == status)
            ):
                return event
        time.sleep(0.5)
    return None


def get_security_alert_items(
    api_client: RESTClient, security_cookie: str
) -> list[dict[str, object]]:
    """Read the security alerts page through the API used by the admin panel."""
    response = api_client.get(
        path="/api/alerts/security?page=1&size=500",
        cookies={ACCESS_COOKIE: security_cookie},
    )
    if response.status != HTTP_OK:
        pytest.fail(f"Security alerts API returned HTTP {response.status}.")
    data = response.body.get("data")
    items = data.get("items") if isinstance(data, dict) else None
    if not isinstance(items, list):
        pytest.fail("Security alerts API returned an invalid items list.")
    return [item for item in items if isinstance(item, dict)]


def wait_for_new_security_alert(
    api_client: RESTClient,
    security_cookie: str,
    category: str,
    *,
    after_rec_id: int,
    timeout: float = 20.0,
) -> dict[str, object] | None:
    """Poll the admin-panel alerts API until a new event appears in one category."""
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        for event in get_security_alert_items(api_client, security_cookie):
            if (
                event.get("Source") == "log-analysis"
                and event.get("Category") == category
                and int(str(event.get("RecID", 0))) > after_rec_id
            ):
                return event
        time.sleep(0.5)
    return None


def test_one_failed_login_is_recorded_as_auth_failure(
    api_client: RESTClient,
    security_cookie: str,
) -> None:
    """Test case: one rejected login creates an auth-failure event.

    Preparation:
    - The Auth Service, Web Gateway, Log Analysis, and Security Logging Manager
      are running, with a system-test security-manager CLI identity.

    Action:
    - Submit one invalid password for the configured system-test account
      through the public login API.

    Expected result:
    - The security-manager CLI returns a Log Analysis `auth_failure` event
      created after this rejected login attempt.

    Requirement: auth-service/FR-002, security-monitoring/FR-001, logging-managers/FR-001.
    """
    test_login = SECURITY_LOGGING_MANAGER_LOGIN
    existing_events = get_security_alert_items(api_client, security_cookie)
    existing_rec_ids = [int(str(event.get("RecID", 0))) for event in existing_events]
    latest_rec_id = max(existing_rec_ids, default=0)
    submit_failed_login(api_client, test_login)
    event = wait_for_new_security_alert(
        api_client,
        security_cookie,
        "auth_failure",
        after_rec_id=latest_rec_id,
    )
    expect((event or {}).get("Category")).to_be("auth_failure")
    expect((event or {}).get("Status")).to_be("active")
    expect((event or {}).get("Acknowledged")).to_be(False)


def test_repeated_failed_logins_create_active_password_guessing_alert(
    api_client: RESTClient,
    security_cli_client: CLIClient,
    security_cookie: str,
) -> None:
    """Test case: repeated rejected logins create an active password-guessing incident.

    Preparation:
    - The Log Analysis password-guessing rule is configured for three failures
      within its detection window, with a three-minute quiet-period cooldown.

    Action:
    - Submit four failed login attempts for the same unique nonexistent login.

    Expected result:
    - The Security Logging Manager's active-unacknowledged list contains the
      new `password_guessing` event, which contributes to the UI counter.

    Requirement: security-monitoring/FR-001, logging-managers/FR-004.
    """
    test_login = new_test_login()
    started_at = datetime.now(UTC).replace(microsecond=0)
    for _ in range(4):
        submit_failed_login(api_client, test_login)

    event = find_security_event(
        security_cli_client,
        "password_guessing",
        created_after=started_at,
        status="active",
        command=ACTIVE_UNACKNOWLEDGED_COMMAND,
    )
    expect(str((event or {}).get("Message", ""))).to_contain("via Web/API")
    count_response = api_client.get(
        path="/api/alerts/security/unacknowledged/count",
        cookies={ACCESS_COOKIE: security_cookie},
    )
    security_count = count_response.body.get("data", {}).get("count", 0)
    expect((event or {}).get("Status")).to_be("active")
    expect((event or {}).get("Acknowledged")).to_be(False)
    expect(count_response.status).to_be(HTTP_OK)
    expect(security_count).to_be_greater_than(0)
