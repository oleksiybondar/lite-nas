"""Installed-system privacy checks for fixed File Service profile resources."""

import base64
import shlex
from collections.abc import Generator
from urllib.parse import urlsplit
from uuid import uuid4

import pytest
from constants import (
    API_BASE_URL,
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
    TEST_SUDO_LOGIN,
)
from hyperiontf import CLIClient, RESTClient, expect

from .conftest import ACCESS_COOKIE

pytestmark = [pytest.mark.api, pytest.mark.FileService, pytest.mark.Nginx]

ORIGIN = f"{urlsplit(API_BASE_URL).scheme}://{urlsplit(API_BASE_URL).netloc}"
PROFILE_DIR = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas"
AVATAR = base64.b64decode(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADElEQVR4nGP4z8AAAAMBAQDJ/pLvAAAAAElFTkSuQmCC"
)


@pytest.fixture
def profile_workspace(operator_cli_client: CLIClient) -> Generator[str, None, None]:
    """Back up fixed profile files and restore the test account after each case."""
    backup_root = (
        f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-profile-backup-{uuid4().hex}"
    )
    names = ("ui.json", "admin-panel.json", "avatar.image")
    operator_cli_client.execute(f"mkdir -m 0700 -- {shlex.quote(backup_root)}")
    operator_cli_client.assert_exit_code(0)
    directory_marker = f"{backup_root}/directory-existed"
    operator_cli_client.execute(
        f"if test -d {shlex.quote(PROFILE_DIR)}; then touch -- {shlex.quote(directory_marker)}; fi"
    )
    operator_cli_client.assert_exit_code(0)
    for name in names:
        original = f"{PROFILE_DIR}/{name}"
        backup = f"{backup_root}/{name}"
        operator_cli_client.execute(
            f"if test -e {shlex.quote(original)} || test -L {shlex.quote(original)}; "
            f"then cp -a -- {shlex.quote(original)} {shlex.quote(backup)}; fi"
        )
        operator_cli_client.assert_exit_code(0)
    try:
        yield PROFILE_DIR
    finally:
        for name in names:
            original = f"{PROFILE_DIR}/{name}"
            backup = f"{backup_root}/{name}"
            operator_cli_client.execute(
                f"if test -e {shlex.quote(backup)} || test -L {shlex.quote(backup)}; "
                f"then mv -f -- {shlex.quote(backup)} {shlex.quote(original)}; "
                f"else rm -f -- {shlex.quote(original)}; fi"
            )
            operator_cli_client.assert_exit_code(0)
        operator_cli_client.execute(
            f"if test ! -e {shlex.quote(directory_marker)}; "
            f"then rmdir -- {shlex.quote(PROFILE_DIR)} 2>/dev/null || true; fi"
        )
        operator_cli_client.execute(f"rm -rf -- {shlex.quote(backup_root)}")


def put_avatar(api_client: RESTClient, cookie: str) -> None:
    """Publish the test owner's complete PNG through the fixed avatar route."""
    response = api_client.put(
        path="/f/profile/avatar",
        payload=AVATAR,
        headers={"Origin": ORIGIN, "Content-Type": "image/png"},
        cookies={ACCESS_COOKIE: cookie},
    )
    if response.status not in (200, 201):
        pytest.fail(f"avatar publication failed: HTTP {response.status}")


# Requirements: file-service-svc/FR-012.
@pytest.mark.parametrize("resource", ["ui", "admin-panel"])
@pytest.mark.usefixtures("profile_workspace")
def test_profile_document_round_trip(
    api_client: RESTClient, operator_cookie: str, resource: str
) -> None:
    """Test case: a user can read their own fixed profile document.

    Preparation:
    - The operator has a valid session; any prior document is backed up.

    Action:
    - PUT a small document and GET the same fixed resource.

    Expected result:
    - The owner's document is returned.
    """
    response = api_client.put(
        path=f"/f/profile/{resource}",
        payload={"document": {"systemTest": resource}},
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    if response.status not in (200, 201):
        pytest.fail(f"profile publication failed: HTTP {response.status}")
    read = api_client.get(path=f"/f/profile/{resource}", cookies={ACCESS_COOKIE: operator_cookie})
    expect(read.body).to_be({"systemTest": resource})


# Requirements: file-service-svc/FR-012.
@pytest.mark.usefixtures("profile_workspace")
def test_profile_expression_text_remains_literal_data(
    api_client: RESTClient, operator_cli_client: CLIClient, operator_cookie: str
) -> None:
    """Test case: profile JSON does not execute expression-shaped text.

    Preparation:
    - The operator's existing profile document is backed up.

    Action:
    - Publish and read a profile document containing template and code syntax
      that would create a separate sentinel if interpreted.

    Expected result:
    - The document is returned literally and no sentinel is created.
    """
    sentinel = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-eval-{uuid4().hex}"
    document = {
        "template": "{{ 7 * 7 }}",
        "expression": f"__import__('os').system('touch {sentinel}')",
    }
    try:
        written = api_client.put(
            path="/f/profile/ui",
            payload={"document": document},
            headers={"Origin": ORIGIN},
            cookies={ACCESS_COOKIE: operator_cookie},
        )
        if written.status not in (200, 201):
            pytest.fail(f"profile publication failed: HTTP {written.status}")
        read = api_client.get(path="/f/profile/ui", cookies={ACCESS_COOKIE: operator_cookie})
        if read.body != document:
            pytest.fail("profile expression was not returned literally")
        operator_cli_client.execute(f"test ! -e {shlex.quote(sentinel)}")
        operator_cli_client.assert_exit_code(0)
    finally:
        operator_cli_client.execute(f"rm -f -- {shlex.quote(sentinel)}")


# Requirements: file-service-svc/FR-012.
@pytest.mark.parametrize("name", ["ui.json", "admin-panel.json"])
def test_other_user_cannot_read_private_profile_document(  # noqa: PLR0913
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    testsudo_cookie: str,
    profile_workspace: str,
    name: str,
) -> None:
    """Test case: the avatar exception does not expose private profile JSON.

    Preparation:
    - The operator publishes a private fixed document in their profile directory.
    - A second account has a separate valid session.

    Action:
    - The second account reads the document through the general /f file route.

    Expected result:
    - Access is denied without returning the document.
    """
    resource = "ui" if name == "ui.json" else "admin-panel"
    written = api_client.put(
        path=f"/f/profile/{resource}",
        payload={"document": {"privateMarker": uuid4().hex}},
        headers={"Origin": ORIGIN},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    if written.status not in (200, 201):
        pytest.fail(f"profile publication failed: HTTP {written.status}")
    operator_cli_client.execute(f"chmod 0600 -- {shlex.quote(profile_workspace + '/' + name)}")
    operator_cli_client.assert_exit_code(0)
    response = api_client.get(
        path=f"/f{profile_workspace}/{name}", cookies={ACCESS_COOKIE: testsudo_cookie}
    )
    expect(response.status).to_be(403)


# Requirements: file-service-svc/FR-012.
def test_other_authenticated_user_can_read_fixed_avatar(
    api_client: RESTClient,
    operator_cookie: str,
    testsudo_cookie: str,
    profile_workspace: str,
) -> None:
    """Test case: a signed-in user can see the owner's fixed public avatar.

    Preparation:
    - The operator publishes a valid avatar in their private home namespace.
    - Another account has an independent valid session.

    Action:
    - The second account requests the fixed avatar path through /f.

    Expected result:
    - The image bytes are returned despite the home directory's private traversal rules.
    """
    put_avatar(api_client, operator_cookie)
    response = api_client.get(
        path=f"/f{profile_workspace}/avatar.image",
        cookies={ACCESS_COOKIE: testsudo_cookie},
    )
    expect(response.status).to_be(200)


# Requirements: file-service-svc/FR-012.
def test_anonymous_user_cannot_read_fixed_avatar(
    api_client: RESTClient, operator_cookie: str, profile_workspace: str
) -> None:
    """Test case: the public avatar exception still requires login.

    Preparation:
    - The operator has published a valid fixed avatar.

    Action:
    - Request its host path through /f without a session.

    Expected result:
    - Authentication denies the request.
    """
    put_avatar(api_client, operator_cookie)
    response = api_client.get(path=f"/f{profile_workspace}/avatar.image")
    expect(response.status).to_be(401)


# Requirements: file-service-svc/FR-012.
def test_private_avatar_does_not_receive_cross_user_exception(
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    testsudo_cookie: str,
    profile_workspace: str,
) -> None:
    """Test case: the avatar exception requires the owner's world-read bit.

    Preparation:
    - The operator publishes a valid avatar and then removes its world-read bit.

    Action:
    - Another account requests the fixed avatar path.

    Expected result:
    - The private avatar is not returned.
    """
    put_avatar(api_client, operator_cookie)
    operator_cli_client.execute(f"chmod 0600 -- {shlex.quote(profile_workspace + '/avatar.image')}")
    operator_cli_client.assert_exit_code(0)
    response = api_client.get(
        path=f"/f{profile_workspace}/avatar.image",
        cookies={ACCESS_COOKIE: testsudo_cookie},
    )
    expect(response.status).to_be(404)


# Requirements: file-service-svc/FR-012.
def test_avatar_exception_is_limited_to_inline_file_route(
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    testsudo_cookie: str,
    profile_workspace: str,
) -> None:
    """Test case: the attachment mount does not bypass private-home traversal.

    Preparation:
    - The operator publishes a valid fixed avatar and denies the second user's
      ordinary file read with a named ACL.

    Action:
    - Another account requests that path through the /d attachment mount.

    Expected result:
    - Ordinary RBAC read authorization denies the attachment request.
    """
    put_avatar(api_client, operator_cookie)
    operator_cli_client.execute(
        f"setfacl -m {shlex.quote('u:' + TEST_SUDO_LOGIN + ':---')} -- "
        f"{shlex.quote(profile_workspace + '/avatar.image')}"
    )
    operator_cli_client.assert_exit_code(0)
    response = api_client.get(
        path=f"/d{profile_workspace}/avatar.image",
        cookies={ACCESS_COOKIE: testsudo_cookie},
    )
    expect(response.status).to_be(403)


# Requirements: file-service-svc/FR-012.
def test_avatar_exception_allows_named_acl_denial(
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    operator_cookie: str,
    testsudo_cookie: str,
    profile_workspace: str,
) -> None:
    """Test case: a world-readable avatar remains available under the explicit exception.

    Preparation:
    - The operator publishes an avatar with its required world-read bit.
    - A named ACL denies the second user ordinary file read permission.

    Action:
    - The second account requests the fixed avatar through /f.

    Expected result:
    - The authenticated avatar exception still returns the image.
    """
    put_avatar(api_client, operator_cookie)
    operator_cli_client.execute(
        f"setfacl -m {shlex.quote('u:' + TEST_SUDO_LOGIN + ':---')} -- "
        f"{shlex.quote(profile_workspace + '/avatar.image')}"
    )
    operator_cli_client.assert_exit_code(0)
    response = api_client.get(
        path=f"/f{profile_workspace}/avatar.image",
        cookies={ACCESS_COOKIE: testsudo_cookie},
    )
    expect(response.status).to_be(200)
