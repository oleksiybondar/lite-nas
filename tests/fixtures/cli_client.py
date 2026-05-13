from collections.abc import Generator

import pytest
from constants import (
    SECURITY_LOGGING_MANAGER_LOGIN,
    SECURITY_LOGGING_MANAGER_PASSWORD,
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
    SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
)
from hyperiontf import CLIClient
from hyperiontf.executors.pytest import (
    fixture,
    hyperion_test_case_setup,  # noqa: F401
)


@fixture(autouse=True, log=False)  # type: ignore[untyped-decorator]
def cli_client(request: pytest.FixtureRequest) -> Generator[CLIClient, None, None]:
    """Create a stable shell-backed CLI client for infrastructure test cases."""
    client = CLIClient()
    client.start_session()
    request.addfinalizer(client.quit)
    yield client


def authenticate_cli_client_as_user(
    cli_client: CLIClient,
    login: str,
    password: str,
) -> None:
    """Switch an existing shell session to the target user with sudo password handling."""
    cli_client.exec_interactive(f"su - {login}")
    cli_client.wait("Password:", timeout=5)
    cli_client.send_keys(password)
    cli_client.wait("$")

    cli_client.detect_action_prompt()


def replace_prompt(
    cli_client: CLIClient,
) -> None:
    """Replace the current CLI prompt with the provided string."""
    cli_client.exec_interactive("export PS1='hyperion$'")
    cli_client.exec_interactive("export PROMPT='hyperion$")
    cli_client.exec_interactive("export COLUMNS=1024")
    cli_client.wait("$", timeout=1)
    cli_client.detect_action_prompt()


@fixture(log=False)  # type: ignore[untyped-decorator]
def operator_cli_client(cli_client: CLIClient) -> Generator[CLIClient, None, None]:
    """Create a CLI session authenticated as the operator system-test user."""
    authenticate_cli_client_as_user(
        cli_client,
        SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
        SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
    )
    yield cli_client


@fixture(log=False)  # type: ignore[untyped-decorator]
def security_cli_client(cli_client: CLIClient) -> Generator[CLIClient, None, None]:
    """Create a CLI session authenticated as the security system-test user."""
    authenticate_cli_client_as_user(
        cli_client,
        SECURITY_LOGGING_MANAGER_LOGIN,
        SECURITY_LOGGING_MANAGER_PASSWORD,
    )
    yield cli_client
