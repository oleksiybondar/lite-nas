"""Installed-system authentication and command-boundary checks for File Service."""

import shlex
from collections.abc import Generator
from http import HTTPStatus
from urllib.parse import quote, urlsplit
from uuid import uuid4

import pytest
from constants import API_BASE_URL, SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN
from hyperiontf import CLIClient, RESTClient, expect

from .conftest import ACCESS_COOKIE

pytestmark = [pytest.mark.api, pytest.mark.FileService, pytest.mark.Nginx]

ORIGIN = f"{urlsplit(API_BASE_URL).scheme}://{urlsplit(API_BASE_URL).netloc}"


@pytest.fixture
def security_path(operator_cli_client: CLIClient) -> Generator[str, None, None]:
    """Reserve an operator-home path and remove only this test's artifacts."""
    path = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-security-{uuid4().hex}"
    try:
        yield path
    finally:
        operator_cli_client.execute(f"rm -rf -- {shlex.quote(path)}")


# Requirements: file-service-svc/FR-001, FR-003, FR-005, FR-010, FR-011, and FR-012.
@pytest.mark.parametrize(
    ("method", "path", "payload"),
    [
        pytest.param("GET", "/f/profile/ui", None, id="profile-ui-read"),
        pytest.param("PUT", "/f/profile/ui", {"document": {}}, id="profile-ui-write"),
        pytest.param("GET", "/f/profile/admin-panel", None, id="profile-admin-read"),
        pytest.param("PUT", "/f/profile/admin-panel", {"document": {}}, id="profile-admin-write"),
        pytest.param("GET", "/f/profile/avatar", None, id="profile-avatar-read"),
        pytest.param("PUT", "/f/profile/avatar", b"image", id="profile-avatar-write"),
        pytest.param("GET", "/f/proprietary/upload-window", None, id="upload-window"),
        pytest.param("POST", "/f/proprietary/upload", {}, id="upload-create"),
        pytest.param("GET", "/f/proprietary/upload/" + "a" * 48, None, id="upload-progress"),
        pytest.param("HEAD", "/f/proprietary/upload/" + "a" * 48, None, id="upload-head"),
        pytest.param("PATCH", "/f/proprietary/upload/" + "a" * 48, b"x", id="upload-chunk"),
        pytest.param("DELETE", "/f/proprietary/upload/" + "a" * 48, None, id="upload-cancel"),
        pytest.param(
            "PUT", "/f/proprietary/upload/" + "a" * 48 + "/completion", None, id="upload-complete"
        ),
        pytest.param("POST", "/f/mkdir", {"path": "/home/unused"}, id="mkdir"),
        pytest.param(
            "POST", "/f/copy", {"source": "/home/unused", "destination": "/home/unused2"}, id="copy"
        ),
        pytest.param(
            "POST", "/f/move", {"source": "/home/unused", "destination": "/home/unused2"}, id="move"
        ),
        pytest.param("GET", "/f/jobs/" + "a" * 48, None, id="job"),
        pytest.param("DELETE", "/f/home/unused", None, id="file-delete"),
    ],
)
def test_anonymous_file_routes_require_login(
    api_client: RESTClient, method: str, path: str, payload: object
) -> None:
    """Test case: a public File Service route rejects an anonymous request.

    Preparation:
    - The installed service is reachable through Nginx without a session.

    Action:
    - Send the selected route and method without an access token.

    Expected result:
    - Authentication fails before the route processes its input.
    """
    response = api_client.request(method=method, path=path, payload=payload).execute()
    expect(response.status).to_be(401)


# Requirements: file-service-svc/FR-001 and IR-001.
@pytest.mark.parametrize("path", ["/f/docs", "/f/openapi.json"])
def test_anonymous_documentation_is_unavailable(api_client: RESTClient, path: str) -> None:
    """Test case: an anonymous caller cannot retrieve File Service documentation.

    Preparation:
    - The installed service may have API documentation enabled or disabled.

    Action:
    - Request the selected documentation resource without a session.

    Expected result:
    - Authentication or the disabled-documentation policy prevents disclosure.
    """
    response = api_client.get(path=path)
    expect(response.status in (HTTPStatus.UNAUTHORIZED, HTTPStatus.NOT_FOUND)).to_be(True)


# Requirements: file-service-svc/FR-001 and FR-003.
@pytest.mark.parametrize("mount", ["/f", "/d", "/f/download"])
@pytest.mark.parametrize("method", ["GET", "HEAD"])
def test_anonymous_download_mounts_require_login(
    api_client: RESTClient, operator_owned_file: str, mount: str, method: str
) -> None:
    """Test case: each download mount denies an anonymous file request.

    Preparation:
    - The operator owns a private file in their home directory.

    Action:
    - Request that file through the selected public mount and method.

    Expected result:
    - The response is 401 before file bytes or metadata are granted.
    """
    response = api_client.request(method=method, path=f"{mount}{operator_owned_file}").execute()
    expect(response.status).to_be(401)


# Requirements: file-service-svc/FR-003.
@pytest.mark.parametrize("mount", ["/f", "/d", "/f/download"])
def test_owner_can_read_each_download_mount(
    api_client: RESTClient, operator_cookie: str, operator_owned_file: str, mount: str
) -> None:
    """Test case: an authorized file can be read through each public download mount.

    Preparation:
    - The operator owns a private file and has a valid session.

    Action:
    - Request its first byte through the selected mount.

    Expected result:
    - The service grants a one-byte partial response.
    """
    response = api_client.get(
        path=f"{mount}{operator_owned_file}",
        headers={"Range": "bytes=0-0"},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(response.status).to_be(HTTPStatus.PARTIAL_CONTENT)


# Requirements: file-service-svc/FR-003.
@pytest.mark.parametrize("mount", ["/f", "/d", "/f/download"])
@pytest.mark.parametrize("method", ["GET", "HEAD"])
def test_other_user_cannot_read_private_file_from_download_mount(
    api_client: RESTClient,
    testsudo_cookie: str,
    operator_owned_file: str,
    mount: str,
    method: str,
) -> None:
    """Test case: each download mount preserves another user's file privacy.

    Preparation:
    - The operator owns a 0600 file and a second user has a separate session.

    Action:
    - The second user requests the file through the selected mount and method.

    Expected result:
    - No file content or HEAD metadata is authorized.
    """
    response = api_client.request(
        method=method,
        path=f"{mount}{operator_owned_file}",
        headers={"Range": "bytes=0-0"},
        cookies={ACCESS_COOKIE: testsudo_cookie},
    ).execute()
    expect(response.status).to_be(HTTPStatus.FORBIDDEN)


# Requirements: file-service-svc/FR-001 and FR-008.
def test_invalid_bearer_cannot_list_home(api_client: RESTClient) -> None:
    """Test case: an invalid bearer token cannot enumerate a home directory.

    Preparation:
    - The operator's home directory exists on the installed system.

    Action:
    - List it with a malformed bearer token.

    Expected result:
    - The response is 401.
    """
    response = api_client.get(
        path=f"/ls/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}",
        headers={"Authorization": "Bearer invalid token"},
    )
    expect(response.status).to_be(401)


# Requirements: file-service-svc/FR-001 and FR-010.
def test_cookie_mutation_rejects_other_origin(
    api_client: RESTClient, operator_cookie: str, security_path: str
) -> None:
    """Test case: another origin cannot create a directory with a browser cookie.

    Preparation:
    - The operator has a valid browser session and an unused home path.

    Action:
    - POST a directory request with an unrelated Origin.

    Expected result:
    - The service rejects the mutation with 403.
    """
    response = api_client.post(
        path="/f/mkdir",
        payload={"path": security_path},
        headers={"Origin": "https://unrelated.example"},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(response.status).to_be(403)


# Requirements: file-service-svc/FR-010 and SR-001.
def test_url_command_substitution_remains_literal(
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    security_path: str,
) -> None:
    """Test case: a URL path cannot execute command substitution during deletion.

    Preparation:
    - The operator owns a disposable file whose literal name contains shell syntax.
    - A distinct sentinel file is absent from the operator's home.

    Action:
    - DELETE the file through /f using a URL-encoded path.

    Expected result:
    - The shell syntax never creates the sentinel file.
    """
    sentinel = f"{security_path}.sentinel"
    # Expand the slash from HOME only if a shell mistakenly interprets the operand.
    target = (
        f"{security_path}$(touch${{IFS}}${{HOME}}${{HOME%%home*}}{sentinel.rsplit('/', 1)[-1]})"
    )
    operator_cli_client.execute(f"printf safe > {shlex.quote(target)}")
    operator_cli_client.assert_exit_code(0)
    try:
        response = api_client.delete(
            path=f"/f{quote(target, safe='/')}",
            headers={"Origin": ORIGIN},
            cookies={ACCESS_COOKIE: operator_cookie},
        )
        operator_cli_client.execute(f"test ! -e {shlex.quote(sentinel)}")
        operator_cli_client.assert_exit_code(0)
        if response.status != HTTPStatus.NO_CONTENT:
            pytest.fail(f"literal URL path did not reach deletion: HTTP {response.status}")
    finally:
        operator_cli_client.execute(f"rm -f -- {shlex.quote(target)} {shlex.quote(sentinel)}")
