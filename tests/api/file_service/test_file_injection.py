"""Installed-system command-operand checks for File Service paths and JSON bodies."""

import shlex
from collections.abc import Generator
from http import HTTPStatus
from urllib.parse import quote, urlsplit
from uuid import uuid4

import pytest
from constants import API_BASE_URL, SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN
from hyperiontf import CLIClient, RESTClient

from .conftest import ACCESS_COOKIE, injection_name, must_finished_job

pytestmark = [pytest.mark.api, pytest.mark.FileService, pytest.mark.Nginx]

ORIGIN = f"{urlsplit(API_BASE_URL).scheme}://{urlsplit(API_BASE_URL).netloc}"


@pytest.fixture
def injection_tree(operator_cli_client: CLIClient) -> Generator[tuple[str, str], None, None]:
    """Create a disposable home tree and remove its separate execution sentinel."""
    root = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-injection-{uuid4().hex}"
    sentinel = f"{root}.sentinel"
    operator_cli_client.execute(f"mkdir -m 0700 -- {shlex.quote(root)}")
    operator_cli_client.assert_exit_code(0)
    try:
        yield root, sentinel
    finally:
        operator_cli_client.execute(f"rm -rf -- {shlex.quote(root)} {shlex.quote(sentinel)}")


def assert_no_sentinel(operator_cli_client: CLIClient, sentinel: str) -> None:
    """Verify that no command embedded in an API operand created the sentinel."""
    operator_cli_client.execute(f"test ! -e {shlex.quote(sentinel)}")
    operator_cli_client.assert_exit_code(0)


# Requirements: file-service-svc/FR-003 and SR-001.
@pytest.mark.parametrize("mount", ["/f", "/d", "/f/download"])
@pytest.mark.parametrize("method", ["GET", "HEAD"])
def test_download_url_treats_shell_syntax_as_filename(  # noqa: PLR0913
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    injection_tree: tuple[str, str],
    mount: str,
    method: str,
) -> None:
    """Test case: download URL syntax stays a literal file operand.

    Preparation:
    - The operator owns a file with shell syntax in its literal name.
    - A distinct sentinel is absent from the operator's home.

    Action:
    - Read that file through the selected mount and method.

    Expected result:
    - The path is served without executing its embedded command.
    """
    root, sentinel = injection_tree
    target = f"{root}/{injection_name(sentinel)}"
    operator_cli_client.execute(f"printf safe > {shlex.quote(target)}")
    operator_cli_client.assert_exit_code(0)
    response = api_client.request(
        method=method,
        path=f"{mount}{quote(target, safe='/')}",
        headers={"Range": "bytes=0-0"},
        cookies={ACCESS_COOKIE: operator_cookie},
    ).execute()
    expected = HTTPStatus.PARTIAL_CONTENT if method == "GET" else HTTPStatus.OK
    if response.status != expected:
        pytest.fail(f"literal download path failed: HTTP {response.status}")
    assert_no_sentinel(operator_cli_client, sentinel)


# Requirements: file-service-svc/FR-008 and SR-001.
def test_listing_url_treats_shell_syntax_as_directory_name(
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    injection_tree: tuple[str, str],
) -> None:
    """Test case: a listing URL cannot execute its directory name.

    Preparation:
    - The operator owns a directory with shell syntax in its literal name.

    Action:
    - List that directory through /ls.

    Expected result:
    - No execution sentinel is created.
    """
    root, sentinel = injection_tree
    target = f"{root}/{injection_name(sentinel)}"
    operator_cli_client.execute(f"mkdir -m 0700 -- {shlex.quote(target)}")
    operator_cli_client.assert_exit_code(0)
    response = api_client.get(
        path=f"/ls{quote(target, safe='/')}", cookies={ACCESS_COOKIE: operator_cookie}
    )
    if response.status != HTTPStatus.OK:
        pytest.fail(f"literal listing path failed: HTTP {response.status}")
    assert_no_sentinel(operator_cli_client, sentinel)


# Requirements: file-service-svc/FR-010, FR-011, and SR-001.
def test_directory_delete_url_treats_shell_syntax_as_directory_name(
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    injection_tree: tuple[str, str],
) -> None:
    """Test case: queued directory deletion keeps its URL path literal.

    Preparation:
    - The operator owns an empty directory with shell syntax in its name.

    Action:
    - DELETE its URL and wait for the removal job to finish.

    Expected result:
    - The user-scoped deletion command does not create the sentinel.
    """
    root, sentinel = injection_tree
    target = f"{root}/{injection_name(sentinel)}"
    operator_cli_client.execute(f"mkdir -m 0700 -- {shlex.quote(target)}")
    operator_cli_client.assert_exit_code(0)
    response = api_client.delete(
        path=f"/f{quote(target, safe='/')}",
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    if response.status != HTTPStatus.ACCEPTED:
        pytest.fail(f"literal directory deletion was not queued: HTTP {response.status}")
    must_finished_job(api_client, operator_cookie, str(response.body["job_id"]))
    assert_no_sentinel(operator_cli_client, sentinel)


# Requirements: file-service-svc/FR-010, FR-011, and SR-001.
@pytest.mark.parametrize(
    ("prefix", "expected"),
    [
        pytest.param("/f/proprietary/upload/", HTTPStatus.NOT_FOUND, id="upload-id"),
        pytest.param("/f/jobs/", HTTPStatus.BAD_REQUEST, id="job-id"),
    ],
)
def test_resource_id_url_rejects_shell_syntax_before_lookup(  # noqa: PLR0913
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    injection_tree: tuple[str, str],
    prefix: str,
    expected: HTTPStatus,
) -> None:
    """Test case: an operation or upload ID cannot become a command operand.

    Preparation:
    - The operator has a valid session and an absent execution sentinel.

    Action:
    - GET a resource ID containing encoded shell syntax.

    Expected result:
    - The route rejects the malformed ID before any command can run.
    """
    _, sentinel = injection_tree
    response = api_client.get(
        path=f"{prefix}{quote(injection_name(sentinel), safe='')}",
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    if response.status != expected:
        pytest.fail(f"malformed resource ID returned HTTP {response.status}")
    assert_no_sentinel(operator_cli_client, sentinel)


# Requirements: file-service-svc/FR-010 and SR-001.
def test_mkdir_json_path_treats_shell_syntax_as_directory_name(
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    injection_tree: tuple[str, str],
) -> None:
    """Test case: the mkdir JSON path remains a literal command operand.

    Preparation:
    - The operator owns an isolated parent directory.

    Action:
    - POST a child path containing shell syntax to /f/mkdir.

    Expected result:
    - Directory creation does not execute the embedded command.
    """
    root, sentinel = injection_tree
    response = api_client.post(
        path="/f/mkdir",
        payload={"path": f"{root}/{injection_name(sentinel)}"},
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    if response.status != HTTPStatus.CREATED:
        pytest.fail(f"literal mkdir path failed: HTTP {response.status}")
    assert_no_sentinel(operator_cli_client, sentinel)


# Requirements: file-service-svc/FR-010 and SR-001.
@pytest.mark.parametrize("operation", ["copy", "move"])
@pytest.mark.parametrize("field", ["source", "destination"])
def test_transfer_json_path_treats_shell_syntax_as_filename(  # noqa: PLR0913
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    injection_tree: tuple[str, str],
    operation: str,
    field: str,
) -> None:
    """Test case: a transfer JSON path stays literal throughout its queued job.

    Preparation:
    - The operator owns a disposable source file and destination directory.

    Action:
    - Submit copy or move with shell syntax in the selected path field.
    - Wait for the queued operation to finish.

    Expected result:
    - No execution sentinel is created.
    """
    root, sentinel = injection_tree
    injected = f"{root}/{injection_name(sentinel)}"
    source = injected if field == "source" else f"{root}/source.txt"
    destination = injected if field == "destination" else f"{root}/destination.txt"
    operator_cli_client.execute(f"printf safe > {shlex.quote(source)}")
    operator_cli_client.assert_exit_code(0)
    response = api_client.post(
        path=f"/f/{operation}",
        payload={"source": source, "destination": destination},
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    if response.status != HTTPStatus.ACCEPTED:
        pytest.fail(f"literal transfer path was not queued: HTTP {response.status}")
    must_finished_job(api_client, operator_cookie, str(response.body["job_id"]))
    assert_no_sentinel(operator_cli_client, sentinel)
