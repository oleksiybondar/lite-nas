"""Installed-system PAM password-change conversations through the gateway."""

import re
import secrets
import subprocess
from collections.abc import Generator
from http import HTTPStatus
from typing import Any, cast

import pytest
from constants import API_BASE_URL, PASSWORD_TEST_LOGIN
from hyperiontf import RESTClient, expect

LOGIN = "/api/auth/login"
START = "/api/auth/password/start"
REPLY = "/api/auth/password/reply"
LOGOUT = "/api/auth/logout"


@pytest.fixture
def password_account() -> Generator[str, None, None]:
    """Give each case a fresh strong password and restore the test account afterward."""
    if PASSWORD_TEST_LOGIN == "root" or not re.fullmatch(  # noqa: S105 - account name
        r"[A-Za-z0-9_][A-Za-z0-9_.-]*", PASSWORD_TEST_LOGIN
    ):
        pytest.fail("LITENAS_PASSWORD_TEST_LOGIN must be a nonroot Unix account name")
    password = new_password()
    set_password(password)
    try:
        yield password
    finally:
        set_password(new_password())


def new_password() -> str:
    """Create a distinct password that satisfies ordinary PAM quality rules."""
    return f"LiteNAS-{secrets.token_hex(12)}-A9!"


def set_password(password: str) -> None:
    """Reset only the dedicated test account through its exact sudoers command."""
    # The fixture validates the account name before it reaches passwd.
    result = subprocess.run(  # noqa: S603
        ["/usr/bin/sudo", "-n", "/usr/bin/passwd", PASSWORD_TEST_LOGIN],
        input=f"{password}\n{password}\n",
        text=True,
        capture_output=True,
        timeout=30,
        check=False,
    )
    if result.returncode:
        pytest.fail(f"Could not reset {PASSWORD_TEST_LOGIN}: {result.stderr.strip()}")


def expire_password() -> None:
    """Force a password update at the account's next PAM login."""
    # The fixture validates the account name before it reaches passwd.
    result = subprocess.run(  # noqa: S603
        ["/usr/bin/sudo", "-n", "/usr/bin/passwd", "-e", PASSWORD_TEST_LOGIN],
        text=True,
        capture_output=True,
        timeout=30,
        check=False,
    )
    if result.returncode:
        pytest.fail(f"Could not expire {PASSWORD_TEST_LOGIN}: {result.stderr.strip()}")


def must_login(client: RESTClient, password: str) -> dict[str, Any]:
    """Submit the dedicated account credentials and return the session data."""
    response = client.post(path=LOGIN, payload={"login": PASSWORD_TEST_LOGIN, "password": password})
    if response.status != HTTPStatus.OK:
        pytest.fail(f"Password-test login failed: HTTP {response.status}")
    return cast("dict[str, Any]", response.body["data"])


def must_start_change(client: RESTClient, password: str) -> dict[str, Any]:
    """Start an on-demand change with the access cookie from a successful login."""
    login = client.post(path=LOGIN, payload={"login": PASSWORD_TEST_LOGIN, "password": password})
    if login.status != HTTPStatus.OK or "lite-nas-at" not in login.cookies:
        pytest.fail(f"Password-test login did not issue an access cookie: HTTP {login.status}")
    cookie = str(login.cookies["lite-nas-at"])
    # HyperionTF omits an empty dict payload; the API contract requires `{}`.
    response = client.session.post(
        f"{API_BASE_URL}{START}",
        data="{}",
        headers={"Content-Type": "application/json"},
        cookies={"lite-nas-at": cookie},
        timeout=30,
    )
    if response.status_code != HTTPStatus.OK:
        pytest.fail(f"Password-change start failed: HTTP {response.status_code}: {response.text}")
    return cast("dict[str, Any]", response.json()["data"])


def must_reply(client: RESTClient, data: dict[str, Any], answer: str) -> dict[str, Any]:
    """Answer the server's current prompt and return its next PAM outcome."""
    conversation = data.get("conversation")
    if not conversation:
        pytest.fail(f"PAM did not provide a conversation: {data.get('status')}")
    response = client.post(
        path=REPLY,
        payload={
            "conversation_id": conversation["id"],
            "sequence": conversation["sequence"],
            "answer": answer,
        },
    )
    if response.status != HTTPStatus.OK:
        pytest.fail(f"Password-change reply failed: HTTP {response.status}")
    return cast("dict[str, Any]", response.body["data"])


def complete_change(client: RESTClient, data: dict[str, Any], old: str, new: str) -> dict[str, Any]:
    """Answer PAM's current, new, and confirmation prompts until it completes."""
    saw_new = False
    for _ in range(8):
        conversation = data.get("conversation")
        if not conversation:
            return data
        prompt = re.sub(r"\s+", " ", conversation["prompt"].lower())
        if "retype" in prompt or "repeat" in prompt or "again" in prompt:
            answer = new
        elif "new" in prompt:
            answer = new
            saw_new = True
        elif "current" in prompt or "old" in prompt or not saw_new:
            answer = old
        else:
            pytest.fail(f"Unexpected PAM password prompt: {conversation['prompt']}")
        data = must_reply(client, data, answer)
    pytest.fail("PAM password conversation exceeded eight prompts")


def must_reach_confirmation(
    client: RESTClient, data: dict[str, Any], old: str, new: str
) -> dict[str, Any]:
    """Answer current and new prompts until PAM requests the confirmation."""
    for _ in range(6):
        conversation = data.get("conversation")
        if not conversation:
            pytest.fail(f"PAM ended before password confirmation: {data.get('status')}")
        prompt = conversation["prompt"].lower()
        if "retype" in prompt or "repeat" in prompt or "again" in prompt:
            return data
        answer = new if "new" in prompt else old
        data = must_reply(client, data, answer)
    pytest.fail("PAM did not request password confirmation")


def must_reach_new_password(client: RESTClient, data: dict[str, Any], old: str) -> dict[str, Any]:
    """Supply any current-password prompts before PAM asks for a new value."""
    for _ in range(4):
        conversation = data.get("conversation")
        if not conversation:
            pytest.fail(f"PAM ended before new-password prompt: {data.get('status')}")
        if "new" in conversation["prompt"].lower():
            return data
        data = must_reply(client, data, old)
    pytest.fail("PAM did not request a new password")


@pytest.mark.Auth
@pytest.mark.WebGateway
@pytest.mark.api
def test_expired_password_change_issues_session(
    api_client: RESTClient, password_account: str
) -> None:
    """Test case: expired-password login completes through PAM.

    Preparation:
    - The dedicated account has a known password and is marked expired.

    Action:
    - Log in and answer PAM's password-change prompts.

    Expected result:
    - The completed conversation authenticates the user.
    """
    expire_password()
    data = must_login(api_client, password_account)
    data = complete_change(api_client, data, password_account, new_password())
    expect(data["authenticated"]).to_be(True)


@pytest.mark.Auth
@pytest.mark.WebGateway
@pytest.mark.api
def test_on_demand_change_allows_new_password_login(
    api_client: RESTClient, password_account: str
) -> None:
    """Test case: a requested password change survives logout.

    Preparation:
    - The dedicated account has a valid session and current password.

    Action:
    - Change its password through PAM, log out, and log in with the new value.

    Expected result:
    - The new password authenticates a fresh session.
    """
    new = new_password()
    data = must_start_change(api_client, password_account)
    complete_change(api_client, data, password_account, new)
    api_client.post(path=LOGOUT, payload={})
    data = must_login(api_client, new)
    expect(data["authenticated"]).to_be(True)


@pytest.mark.Auth
@pytest.mark.WebGateway
@pytest.mark.api
def test_password_mismatch_does_not_complete_change(
    api_client: RESTClient, password_account: str
) -> None:
    """Test case: mismatched confirmation does not change the password.

    Preparation:
    - The dedicated account has a known current password.

    Action:
    - Start an on-demand change and give a different confirmation.

    Expected result:
    - PAM keeps the change pending or denies it without authentication.
    """
    data = must_start_change(api_client, password_account)
    data = must_reach_confirmation(api_client, data, password_account, new_password())
    data = must_reply(api_client, data, new_password())
    expect(data["authenticated"]).to_be(False)


@pytest.mark.Auth
@pytest.mark.WebGateway
@pytest.mark.api
def test_invalid_current_password_is_denied(api_client: RESTClient, password_account: str) -> None:
    """Test case: PAM rejects the wrong current password.

    Preparation:
    - The dedicated account has a valid password and session.

    Action:
    - Start an on-demand change and answer its current-password prompt incorrectly.

    Expected result:
    - The password-change conversation is denied.
    """
    data = must_start_change(api_client, password_account)
    data = must_reply(api_client, data, new_password())
    expect(data["status"]).to_be("denied")


@pytest.mark.Auth
@pytest.mark.WebGateway
@pytest.mark.api
def test_weak_new_password_is_rejected(api_client: RESTClient, password_account: str) -> None:
    """Test case: PAM password quality rejects a weak new value.

    Preparation:
    - The dedicated account has a valid current password.

    Action:
    - Start an on-demand change and submit a short new password.

    Expected result:
    - PAM returns an error for the rejected value.
    """
    data = must_start_change(api_client, password_account)
    data = must_reach_new_password(api_client, data, password_account)
    data = must_reply(api_client, data, "abc")
    messages = data.get("conversation", {}).get("messages", [])
    expect(any(message["level"] == "error" for message in messages)).to_be(True)


@pytest.mark.Auth
@pytest.mark.WebGateway
@pytest.mark.api
def test_abandoned_expired_change_is_required_on_next_login(
    api_client: RESTClient, password_account: str
) -> None:
    """Test case: abandoning an expired-password conversation keeps expiry.

    Preparation:
    - The dedicated account's password is expired.

    Action:
    - Start login, leave its PAM prompt unanswered, then start a new login.

    Expected result:
    - The next login still requires a password change.
    """
    expire_password()
    must_login(api_client, password_account)
    data = must_login(api_client, password_account)
    expect(data["status"]).to_be("password_change_required")
