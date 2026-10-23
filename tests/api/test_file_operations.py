"""Installed-system authorization checks for File Service mutations and jobs."""

import shlex
import time
from collections.abc import Generator
from http import HTTPStatus
from urllib.parse import urlsplit
from uuid import uuid4

import pytest
from constants import API_BASE_URL, SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN
from hyperiontf import CLIClient, RESTClient, expect

from .conftest import ACCESS_COOKIE

pytestmark = [pytest.mark.api, pytest.mark.FileService, pytest.mark.Nginx]

ORIGIN = f"{urlsplit(API_BASE_URL).scheme}://{urlsplit(API_BASE_URL).netloc}"


@pytest.fixture
def operation_tree(operator_cli_client: CLIClient) -> Generator[tuple[str, str], None, None]:
    """Create and later remove a small private operator-owned operation tree."""
    root = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-operation-{uuid4().hex}"
    source = f"{root}/source.txt"
    operator_cli_client.execute(
        f"mkdir -m 0700 -- {shlex.quote(root)} && "
        f"printf 'operation-source\\n' > {shlex.quote(source)} && "
        f"chmod 0600 -- {shlex.quote(source)}"
    )
    operator_cli_client.assert_exit_code(0)
    try:
        yield root, source
    finally:
        operator_cli_client.execute(f"rm -rf -- {shlex.quote(root)}")


def must_finished_job(api_client: RESTClient, cookie: str, job_id: str) -> None:
    """Wait for a submitted short operation or fail with its reported state."""
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


# Requirements: file-service-svc/FR-010.
def test_mkdir_creates_operator_directory(
    api_client: RESTClient, operator_cookie: str, operation_tree: tuple[str, str]
) -> None:
    """Test case: an authorized user can create one private directory.

    Preparation:
    - The operator owns an isolated test tree.

    Action:
    - POST a private child path to /f/mkdir.

    Expected result:
    - The service creates the requested directory.
    """
    root, _ = operation_tree
    path = f"{root}/created"
    response = api_client.post(
        path="/f/mkdir",
        payload={"path": path, "private": True},
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(response.status).to_be(201)


# Requirements: file-service-svc/FR-010 and FR-011.
@pytest.mark.parametrize("operation", ["copy", "move"])
def test_transfer_job_publishes_authorized_file(
    api_client: RESTClient,
    operator_cookie: str,
    operation_tree: tuple[str, str],
    operation: str,
) -> None:
    """Test case: an authorized transfer job publishes the requested file.

    Preparation:
    - The operator owns a small source file in a disposable tree.

    Action:
    - Start the selected copy or move job and wait for its completion.

    Expected result:
    - The destination serves the source content.
    """
    root, source = operation_tree
    destination = f"{root}/{operation}.txt"
    response = api_client.post(
        path=f"/f/{operation}",
        payload={"source": source, "destination": destination},
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    if response.status != HTTPStatus.ACCEPTED:
        pytest.fail(f"{operation} was not queued: HTTP {response.status}")
    must_finished_job(api_client, operator_cookie, str(response.body["job_id"]))
    downloaded = api_client.get(path=f"/f{destination}", cookies={ACCESS_COOKIE: operator_cookie})
    expect(downloaded.raw_body).to_be("operation-source\n")


# Requirements: file-service-svc/FR-011.
def test_other_user_cannot_read_operation_job(
    api_client: RESTClient,
    operator_cookie: str,
    testsudo_cookie: str,
    operation_tree: tuple[str, str],
) -> None:
    """Test case: a queued job remains visible only to its submitting user.

    Preparation:
    - The operator submits a copy job; another user has a separate session.

    Action:
    - The second account requests the operator's job ID.

    Expected result:
    - The job is hidden with a not-found response.
    """
    root, source = operation_tree
    response = api_client.post(
        path="/f/copy",
        payload={"source": source, "destination": f"{root}/job-owner.txt"},
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    if response.status != HTTPStatus.ACCEPTED:
        pytest.fail(f"copy was not queued: HTTP {response.status}")
    denied = api_client.get(
        path=f"/f/jobs/{response.body['job_id']}",
        cookies={ACCESS_COOKIE: testsudo_cookie},
    )
    must_finished_job(api_client, operator_cookie, str(response.body["job_id"]))
    expect(denied.status).to_be(404)


# Requirements: file-service-svc/FR-010.
def test_other_user_cannot_delete_operator_file(
    api_client: RESTClient,
    testsudo_cookie: str,
    operation_tree: tuple[str, str],
) -> None:
    """Test case: another user cannot delete the operator's private file.

    Preparation:
    - The operator owns a private file in an isolated directory.

    Action:
    - A second account sends DELETE through /f with its own session.

    Expected result:
    - The mutation is denied.
    """
    _, source = operation_tree
    response = api_client.delete(
        path=f"/f{source}",
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: testsudo_cookie},
    )
    expect(response.status).to_be(403)


# Requirements: file-service-svc/FR-010.
def test_operator_can_delete_owned_file(
    api_client: RESTClient,
    operator_cookie: str,
    operation_tree: tuple[str, str],
) -> None:
    """Test case: an owner can delete a regular file through /f.

    Preparation:
    - The operator owns a private disposable file.

    Action:
    - DELETE the file with the operator's session.

    Expected result:
    - The service reports successful deletion.
    """
    _, source = operation_tree
    response = api_client.delete(
        path=f"/f{source}",
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(response.status).to_be(204)
