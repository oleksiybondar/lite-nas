from collections.abc import Generator

import pytest
from hyperiontf import CLIClient
from hyperiontf.executors.pytest import (
    fixture,
    hyperion_test_case_setup,  # noqa: F401
)


@fixture(autouse=True, log=False)  # type: ignore[untyped-decorator]
def cli_client(request: pytest.FixtureRequest) -> Generator[CLIClient, None, None]:
    """Create a stable shell-backed CLI client for infrastructure test cases."""
    # NOTE:
    # Using `sh` instead of `bash` because the CLI client detects the shell
    # prompt dynamically and waits for the exact prompt string to reappear
    # after command execution.
    #
    # `bash` may render shortened/truncated prompts depending on terminal
    # width, hostname, PS1 configuration, or interactive shell behavior
    # (example: `<nas/tests$` instead of the full prompt path).
    #
    # This causes prompt detection and command completion checks to fail.
    #
    # `sh` provides a simpler and more stable prompt format, which avoids
    # prompt truncation issues and keeps the CLI client behavior predictable
    # in automated test environments.
    client = CLIClient("sh")  # using sh,not bash because the client
    client.start_session()
    request.addfinalizer(client.quit)
    yield client
