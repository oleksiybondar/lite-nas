"""Reuse installed accounts and a media file with explicit service-read ACL."""

import shlex
from collections.abc import Generator
from uuid import uuid4

import pytest
from api.file_service.conftest import operator_cookie, operator_owned_file, testsudo_cookie
from constants import SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN
from hyperiontf import CLIClient

__all__ = ["operator_cookie", "operator_owned_file", "testsudo_cookie"]


@pytest.fixture
def readable_media(operator_cli_client: CLIClient) -> Generator[str, None, None]:
    """Create an operator-owned AVI whose service group may read its bytes."""
    path = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-stream-{uuid4().hex}.avi"
    quoted = shlex.quote(path)
    operator_cli_client.execute(
        f"printf 'sample-video-bytes' > {quoted} && chmod 0600 {quoted} "
        f"&& setfacl -m g:lite-nas-fs-rd:r {quoted}"
    )
    operator_cli_client.assert_exit_code(0)
    try:
        yield path
    finally:
        operator_cli_client.execute(f"rm -f -- {quoted}")


@pytest.fixture
def valid_avi(operator_cli_client: CLIClient) -> Generator[str, None, None]:
    """Generate a short private AVI for the installed OS transcoder check."""
    path = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-transcode-{uuid4().hex}.avi"
    quoted = shlex.quote(path)
    operator_cli_client.execute(
        "ffmpeg -nostdin -loglevel error -y -f lavfi "
        "-i color=c=blue:s=64x64:r=1 -t 1 -c:v mpeg4 -f avi "
        f"{quoted} && chmod 0600 {quoted}"
    )
    operator_cli_client.assert_exit_code(0)
    try:
        yield path
    finally:
        operator_cli_client.execute(f"rm -f -- {quoted}")
