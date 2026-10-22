"""Shared installed-system File Service account and file fixtures."""

import os
import shlex
from collections.abc import Generator
from http import HTTPStatus
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


def login_cookie(api_client: RESTClient, login: str, password: str) -> str:
    """Log in through the browser API and return the access cookie value."""
    response = api_client.post(path=LOGIN_ENDPOINT, payload={"login": login, "password": password})
    if response.status != HTTPStatus.OK or ACCESS_COOKIE not in response.cookies:
        pytest.fail(f"login did not issue {ACCESS_COOKIE}: HTTP {response.status}")
    return str(response.cookies[ACCESS_COOKIE])


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
