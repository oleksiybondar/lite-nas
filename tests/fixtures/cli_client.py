import os
import re
from collections.abc import Generator

import pytest
from constants import (
    SECURITY_LOGGING_MANAGER_LOGIN,
    SECURITY_LOGGING_MANAGER_PASSWORD,
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
    SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
    TEST_SUDO_LOGIN,
    TEST_SUDO_PASSWORD,
)
from hyperiontf import CLIClient
from hyperiontf.executors.pytest import hyperion_test_case_setup  # noqa: F401

CLI_DEFAULT_PROMPT_ENV = {
    "PS1": "litenas:hyperion$",
    "PROMPT": "litenas:hyperion$",
    "COLUMNS": "1024",
}
CLI_PAGER_ENV = {
    "PAGER": "cat",
    "MANPAGER": "cat",
    "SYSTEMD_PAGER": "cat",
    "GIT_PAGER": "cat",
    "APT_PAGER": "cat",
}
CLI_DEFAULT_ARGS = ["--noprofile", "--norc"]
CLI_DEFAULT_ENV = {
    "NO_COLOR": "1",
    "CLICOLOR": "0",
    "CLICOLOR_FORCE": "0",
    "FORCE_COLOR": "0",
    "TERM": "dumb",
    **CLI_DEFAULT_PROMPT_ENV,
    **CLI_PAGER_ENV,
}


@pytest.fixture(autouse=True)
def cli_client() -> Generator[CLIClient, None, None]:
    """Create a stable shell-backed CLI client for infrastructure test cases."""
    client = CLIClient(
        shell_args=CLI_DEFAULT_ARGS,
        env=CLI_DEFAULT_ENV,
    )
    # HyperionTF 0.7.4 drops chunks equal to action_prompt. Match a unique suffix
    # so a separately delivered prompt still completes delayed commands.
    client.action_prompt = "hyperion$"
    try:
        yield client
    finally:
        client.quit()


def authenticate_cli_client_as_user(
    cli_client: CLIClient,
    login: str,
    password: str,
) -> None:
    """Switch to a system-test user, supporting both regular and root runners."""
    if not re.fullmatch(r"[a-zA-Z0-9_][a-zA-Z0-9_.-]*", login):
        raise ValueError
    cli_client.exec_interactive(f"su - {login}")
    if os.geteuid() != 0:
        cli_client.wait("Password:", timeout=5)
        cli_client.send_keys(password)
    cli_client.wait("$", timeout=10)
    replace_prompt(cli_client)


def replace_prompt(
    cli_client: CLIClient,
) -> None:
    """Restore deterministic prompts and disable pagers after an account switch."""
    for key, value in {**CLI_DEFAULT_PROMPT_ENV, **CLI_PAGER_ENV}.items():
        cli_client.exec_interactive(f"export {key}={value!r}")
    cli_client.wait("$")
    cli_client.detect_action_prompt()
    cli_client.action_prompt = "hyperion$"


@pytest.fixture
def operator_cli_client(cli_client: CLIClient) -> CLIClient:
    """Create a CLI session authenticated as the operator system-test user."""
    authenticate_cli_client_as_user(
        cli_client,
        SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
        SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
    )
    return cli_client


@pytest.fixture
def security_cli_client(cli_client: CLIClient) -> CLIClient:
    """Create a CLI session authenticated as the security system-test user."""
    authenticate_cli_client_as_user(
        cli_client,
        SECURITY_LOGGING_MANAGER_LOGIN,
        SECURITY_LOGGING_MANAGER_PASSWORD,
    )
    return cli_client


@pytest.fixture
def testsudo_cli_client() -> Generator[CLIClient, None, None]:
    """Create an independent shell authenticated as the restricted sudo test user."""
    client = CLIClient(
        shell_args=CLI_DEFAULT_ARGS,
        env=CLI_DEFAULT_ENV,
    )
    client.action_prompt = "hyperion$"
    try:
        authenticate_cli_client_as_user(
            client,
            TEST_SUDO_LOGIN,
            TEST_SUDO_PASSWORD,
        )
        yield client
    finally:
        client.quit()
