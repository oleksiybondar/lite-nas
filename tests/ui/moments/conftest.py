"""Synthetic HEIC media fixture for installed Moments browser tests."""

import os
import shlex
from collections.abc import Generator
from pathlib import Path
from tempfile import NamedTemporaryFile
from uuid import uuid4

import pytest
from constants import SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN
from hyperiontf import CLIClient


@pytest.fixture
def operator_heic(operator_cli_client: CLIClient) -> Generator[str, None, None]:
    """Create a private operator-owned copy of a synthetic 64-pixel HEIC."""
    source = Path(__file__).parents[2] / "fixtures/media/blue-64.HEIC"
    with NamedTemporaryFile(prefix="lite-nas-heic-ui-", suffix=".HEIC", delete=False) as staging:
        staging.write(source.read_bytes())
        staging_path = staging.name
    os.chmod(staging_path, 0o644)
    path = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-ui-{uuid4().hex}.HEIC"
    try:
        operator_cli_client.execute(
            f"cp -- {shlex.quote(staging_path)} {shlex.quote(path)} "
            f"&& chmod 0600 {shlex.quote(path)}"
        )
        operator_cli_client.assert_exit_code(0)
        yield path
    finally:
        operator_cli_client.execute(f"rm -f -- {shlex.quote(path)}")
        Path(staging_path).unlink(missing_ok=True)
