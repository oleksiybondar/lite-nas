"""Authenticated Moments API and disposable gallery fixtures."""

import shlex
from collections.abc import Generator
from http import HTTPStatus
from uuid import uuid4

import pytest
from api.file_service.conftest import ACCESS_COOKIE
from constants import SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN
from hyperiontf import CLIClient, RESTClient


@pytest.fixture
def moments_source(operator_cli_client: CLIClient) -> Generator[tuple[str, str], None, None]:
    """Create one operator-owned directory and file, then remove both after use."""
    directory = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-moments-{uuid4().hex}"
    file_path = f"{directory}/photo.jpg"
    operator_cli_client.execute(
        f"mkdir -m 0700 -- {shlex.quote(directory)} "
        f"&& printf 'fixture' > {shlex.quote(file_path)}"
    )
    operator_cli_client.assert_exit_code(0)
    try:
        yield directory, file_path
    finally:
        operator_cli_client.execute(f"rm -rf -- {shlex.quote(directory)}")


@pytest.fixture
def gallery_ids(
    moments_client: RESTClient, moments_cookie: str
) -> Generator[list[str], None, None]:
    """Track galleries created by a test and delete their metadata on teardown."""
    ids: list[str] = []
    yield ids
    for gallery_id in ids:
        response = moments_client.delete(
            path=f"/api/galleries/{gallery_id}", cookies={ACCESS_COOKIE: moments_cookie}
        )
        if response.status not in (HTTPStatus.NO_CONTENT, HTTPStatus.NOT_FOUND):
            pytest.fail(f"gallery cleanup failed: HTTP {response.status}")


def must_create_gallery(  # noqa: PLR0913
    client: RESTClient,
    cookie: str,
    gallery_ids: list[str],
    kind: str,
    name: str,
    directory: str | None = None,
) -> str:
    """Create an album or collection and register its metadata for cleanup."""
    payload = {"type": kind, "name": name}
    if directory is not None:
        payload["uploadDirectory"] = directory
    response = client.post(path="/api/galleries", payload=payload, cookies={ACCESS_COOKIE: cookie})
    if response.status != HTTPStatus.OK:
        pytest.fail(f"gallery creation failed: HTTP {response.status}: {response.body}")
    gallery_id = str(response.body["id"])
    gallery_ids.append(gallery_id)
    return gallery_id
