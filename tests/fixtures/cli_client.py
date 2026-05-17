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

CLI_DEFAULT_PROMPT_ENV = {"PS1": "hyperion$", "PROMPT": "hyperion$", "COLUMNS": "1024"}
CLI_DEFAULT_ARGS = ["--noprofile", "--norc"]
CLI_DEFAULT_ENV = {
    "NO_COLOR": "1",
    "CLICOLOR": "0",
    "CLICOLOR_FORCE": "0",
    "FORCE_COLOR": "0",
    "TERM": "dumb",
    **CLI_DEFAULT_PROMPT_ENV,
}


@fixture(autouse=True, log=False)  # type: ignore[untyped-decorator]
def cli_client(request: pytest.FixtureRequest) -> Generator[CLIClient, None, None]:
    """Create a stable shell-backed CLI client for infrastructure test cases."""
    client = CLIClient(
        shell_args=CLI_DEFAULT_ARGS,
        env={
            "NO_COLOR": "1",
            "CLICOLOR": "0",
            "CLICOLOR_FORCE": "0",
            "FORCE_COLOR": "0",
            "TERM": "dumb",
            "PS1": "hyperion$",
            "PROMPT": "hyperion$",
            "COLUMNS": "1024",
        },
    )
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
    replace_prompt(cli_client)


def replace_prompt(
    cli_client: CLIClient,
) -> None:
    """Replace the current CLI prompt with the provided string."""
    for key, value in CLI_DEFAULT_PROMPT_ENV.items():
        cli_client.exec_interactive(f"export {key}='{value}'")
    cli_client.wait("$")
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
