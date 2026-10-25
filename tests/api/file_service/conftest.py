"""Shared installed-system File Service account and file fixtures."""

import os
import shlex
import time
from collections.abc import Generator
from http import HTTPStatus
from pathlib import PurePosixPath
from uuid import uuid4

import pytest
from constants import (
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
    SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
    TEST_SUDO_LOGIN,
    TEST_SUDO_PASSWORD,
)
from hyperiontf import CLIClient, RESTClient

ACCESS_COOKIE = "lite-nas-at"
LOGIN_ENDPOINT = "/api/auth/login"


def injection_name(sentinel: str) -> str:
    """Return a literal path segment that would touch sentinel if sent to a shell."""
    name = PurePosixPath(sentinel).name
    # HOME%%home* expands to a slash only when a shell interprets the operand.
    return f"$(touch${{IFS}}${{HOME}}${{HOME%%home*}}{name})"


def login_cookie(api_client: RESTClient, login: str, password: str) -> str:
    """Return an access cookie without retaining it in the shared REST session."""
    response = api_client.post(path=LOGIN_ENDPOINT, payload={"login": login, "password": password})
    if response.status != HTTPStatus.OK or ACCESS_COOKIE not in response.cookies:
        pytest.fail(f"login did not issue {ACCESS_COOKIE}: HTTP {response.status}")
    cookie = str(response.cookies[ACCESS_COOKIE])
    api_client.session.cookies.clear()
    return cookie


def must_finished_job(api_client: RESTClient, cookie: str, job_id: str) -> None:
    """Wait for a short File Service job or fail with its reported error."""
    deadline = time.monotonic() + 20
    while time.monotonic() < deadline:
        response = api_client.get(path=f"/f/jobs/{job_id}", cookies={ACCESS_COOKIE: cookie})
        if response.status != HTTPStatus.OK:
            pytest.fail(f"job lookup failed: HTTP {response.status}")
        if response.body["state"] == "done":
            return
        if response.body["state"] == "error":
            pytest.fail(f"job failed: {response.body.get('error_code')}")
        time.sleep(0.1)
    pytest.fail("job did not finish within 20 seconds")


@pytest.fixture
def operator_cookie(api_client: RESTClient) -> str:
    """Log in as the installed operator through the browser API."""
    return login_cookie(
        api_client,
        SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
        SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
    )


@pytest.fixture
def testsudo_cookie(api_client: RESTClient) -> str:
    """Log in as the installed restricted sudo test user through the browser API."""
    return login_cookie(api_client, TEST_SUDO_LOGIN, TEST_SUDO_PASSWORD)


@pytest.fixture
def admin_cookie(api_client: RESTClient) -> str:
    """Require and log in as the explicitly configured File Service admin."""
    login = os.getenv("LITENAS_FILE_SERVICE_ADMIN_LOGIN")
    password = os.getenv("LITENAS_FILE_SERVICE_ADMIN_PASSWORD")
    if not login or not password:
        pytest.fail(
            "Set LITENAS_FILE_SERVICE_ADMIN_LOGIN and "
            "LITENAS_FILE_SERVICE_ADMIN_PASSWORD for File Service admin tests"
        )
    return login_cookie(api_client, login, password)


@pytest.fixture
def operator_owned_file(operator_cli_client: CLIClient) -> Generator[str, None, None]:
    """Create a private operator-owned file and remove it after the test."""
    path = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-file-test-{uuid4().hex}"
    quoted = shlex.quote(path)
    operator_cli_client.execute(
        f"printf 'operator-owned-file\\n' > {quoted} && chmod 0600 {quoted}"
    )
    operator_cli_client.assert_exit_code(0)
    try:
        yield path
    finally:
        operator_cli_client.execute(f"rm -f -- {quoted}")
