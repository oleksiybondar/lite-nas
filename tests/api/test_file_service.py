"""Installed-system file access checks through Nginx's public mounts."""

import pytest
from constants import (
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
)
from hyperiontf import RESTClient, expect

from .conftest import ACCESS_COOKIE

pytestmark = [pytest.mark.api, pytest.mark.FileService, pytest.mark.Nginx]

# Add paths here as installed-system coverage grows. The table deliberately
# includes world-readable /etc/passwd: operator policy must still deny it.
# Use a one-byte Range so a failed check never logs a whole sensitive file.
CRITICAL_FILE_PATHS = [
    pytest.param("/etc/passwd", id="public-system-account-file"),
    pytest.param("/etc/shadow", id="private-system-account-file"),
    pytest.param("/etc/lite-nas/auth.conf", id="litenas-auth-configuration"),
    pytest.param("/var/log/lite-nas/auth-service.log", id="litenas-auth-log"),
]

ANONYMOUS_LIST_PATHS = [
    pytest.param("/etc", id="system-configuration-directory"),
    pytest.param("/var/log", id="system-log-directory"),
    pytest.param(f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}", id="operator-home-directory"),
]

OPERATOR_DENIED_LIST_PATHS = [
    pytest.param("/etc", id="system-configuration-directory"),
    pytest.param("/var/log", id="system-log-directory"),
]


# Requirements: file-service-svc/FR-001 and file-service-svc/FR-003.
@pytest.mark.parametrize("host_path", CRITICAL_FILE_PATHS)
def test_anonymous_file_read_is_denied(api_client: RESTClient, host_path: str) -> None:
    """Test case: anonymous callers cannot download host files.

    Preparation:
    - File Service and Nginx are running with installed system paths.

    Action:
    - Request a one-byte range through the public /f mount without a token.

    Expected result:
    - File Service returns 401 before granting file bytes.
    """
    response = api_client.get(path=f"/f{host_path}", headers={"Range": "bytes=0-0"})
    expect(response.status).to_be(401)


# Requirements: file-service-svc/FR-001 and file-service-svc/FR-008.
@pytest.mark.parametrize("host_path", ANONYMOUS_LIST_PATHS)
def test_anonymous_directory_listing_is_denied(api_client: RESTClient, host_path: str) -> None:
    """Test case: anonymous callers cannot list host directories.

    Preparation:
    - File Service and Nginx are running with installed system paths.

    Action:
    - Request a directory listing through the public /ls mount without a token.

    Expected result:
    - File Service returns 401 before exposing directory entries.
    """
    response = api_client.get(path=f"/ls{host_path}")
    expect(response.status).to_be(401)


# Requirements: file-service-svc/FR-001 and file-service-svc/FR-003.
def test_anonymous_operator_file_read_is_denied(
    api_client: RESTClient, operator_owned_file: str
) -> None:
    """Test case: anonymous callers cannot read an operator's private file.

    Preparation:
    - The operator owns a private file in their home directory.

    Action:
    - Request it through Nginx's /f mount without a token.

    Expected result:
    - File Service returns 401 before granting file bytes.
    """
    response = api_client.get(path=f"/f{operator_owned_file}", headers={"Range": "bytes=0-0"})
    expect(response.status).to_be(401)


# Requirements: file-service-svc/FR-001 and file-service-svc/IR-001.
@pytest.mark.usefixtures("operator_cookie")
def test_login_cookie_reaches_file_service(api_client: RESTClient) -> None:
    """Test case: a browser-style login cookie reaches File Service through Nginx.

    Preparation:
    - The operator logs in through the admin-panel API on the same origin.
    - The REST client retains the issued access cookie in its session.

    Action:
    - Request the File Service upload window through Nginx without manually
      adding a Cookie or Authorization header.

    Expected result:
    - File Service authenticates the session and returns the window resource.
    """
    response = api_client.get(path="/f/proprietary/upload-window")
    expect(response.status).to_be(200)


# Requirements: file-service-svc/FR-003.
def test_operator_can_read_own_file(
    api_client: RESTClient, operator_cookie: str, operator_owned_file: str
) -> None:
    """Test case: an authenticated operator can read their own file.

    Preparation:
    - The operator owns a private file and has logged in through the admin-panel API.

    Action:
    - Request the file through Nginx's /f mount with the login cookie.

    Expected result:
    - File Service returns the operator-owned content.
    """
    response = api_client.get(
        path=f"/f{operator_owned_file}", cookies={ACCESS_COOKIE: operator_cookie}
    )
    expect(response.raw_body).to_be("operator-owned-file\n")


# Requirements: file-service-svc/FR-008.
def test_operator_can_list_own_file(
    api_client: RESTClient, operator_cookie: str, operator_owned_file: str
) -> None:
    """Test case: an operator can find their own file in a directory listing.

    Preparation:
    - The operator owns a private file and has logged in through the admin-panel API.

    Action:
    - List the operator's home directory through Nginx's /ls mount.

    Expected result:
    - The listing contains the operator-owned file.
    """
    response = api_client.get(
        path=f"/ls/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}",
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(operator_owned_file in response.body["files"]).to_be(True)


# Requirements: file-service-svc/FR-003.
@pytest.mark.parametrize("host_path", CRITICAL_FILE_PATHS)
def test_operator_cannot_read_critical_file(
    api_client: RESTClient, operator_cookie: str, host_path: str
) -> None:
    """Test case: an operator cannot download critical system data.

    Preparation:
    - The operator has logged in through the admin-panel API.
    - The installed system provides the parametrized critical file.

    Action:
    - Request a one-byte range through Nginx's /f mount.

    Expected result:
    - File Service returns 403 before granting file bytes.
    """
    response = api_client.get(
        path=f"/f{host_path}",
        headers={"Range": "bytes=0-0"},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(response.status).to_be(403)


# Requirements: file-service-svc/FR-008.
@pytest.mark.parametrize("host_path", OPERATOR_DENIED_LIST_PATHS)
def test_operator_cannot_list_critical_directory(
    api_client: RESTClient, operator_cookie: str, host_path: str
) -> None:
    """Test case: an operator cannot enumerate critical system directories.

    Preparation:
    - The operator has logged in through the admin-panel API.

    Action:
    - Request a critical directory through Nginx's /ls mount.

    Expected result:
    - File Service returns 403 before exposing entries.
    """
    response = api_client.get(path=f"/ls{host_path}", cookies={ACCESS_COOKIE: operator_cookie})
    expect(response.status).to_be(403)


# Requirements: file-service-svc/FR-003.
@pytest.mark.parametrize("host_path", CRITICAL_FILE_PATHS)
def test_admin_can_read_critical_file(
    api_client: RESTClient, admin_cookie: str, host_path: str
) -> None:
    """Test case: the configured admin can read critical system files.

    Preparation:
    - The configured admin account has logged in through the admin-panel API.
    - The installed system provides the parametrized critical file.

    Action:
    - Request one byte through Nginx's /f mount.

    Expected result:
    - File Service returns partial content for the requested file.
    """
    response = api_client.get(
        path=f"/f{host_path}",
        headers={"Range": "bytes=0-0"},
        cookies={ACCESS_COOKIE: admin_cookie},
    )
    expect(response.status).to_be(206)


# Requirements: file-service-svc/FR-008.
@pytest.mark.parametrize("host_path", OPERATOR_DENIED_LIST_PATHS)
def test_admin_can_list_critical_directory(
    api_client: RESTClient, admin_cookie: str, host_path: str
) -> None:
    """Test case: the configured admin can list critical directories.

    Preparation:
    - The configured admin account has logged in through the admin-panel API.

    Action:
    - Request a critical directory through Nginx's /ls mount.

    Expected result:
    - File Service returns a successful listing response.
    """
    response = api_client.get(path=f"/ls{host_path}", cookies={ACCESS_COOKIE: admin_cookie})
    expect(response.status).to_be(200)


# Requirements: file-service-svc/FR-003.
def test_admin_can_read_operator_file(
    api_client: RESTClient, admin_cookie: str, operator_owned_file: str
) -> None:
    """Test case: the configured admin can read another user's private file.

    Preparation:
    - The operator owns a private file and the admin has logged in.

    Action:
    - Request one byte through Nginx's /f mount with the admin cookie.

    Expected result:
    - File Service returns partial content for the operator-owned file.
    """
    response = api_client.get(
        path=f"/f{operator_owned_file}",
        headers={"Range": "bytes=0-0"},
        cookies={ACCESS_COOKIE: admin_cookie},
    )
    expect(response.status).to_be(206)
