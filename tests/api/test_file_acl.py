"""Installed-system POSIX ACL checks through File Service and RBAC."""

import shlex
from collections.abc import Generator
from uuid import uuid4

import pytest
from constants import SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN, TEST_SUDO_LOGIN
from hyperiontf import CLIClient, RESTClient, expect

from .conftest import ACCESS_COOKIE

pytestmark = [pytest.mark.api, pytest.mark.FileService, pytest.mark.Nginx, pytest.mark.ACL]


@pytest.fixture
def acl_file(operator_cli_client: CLIClient) -> Generator[str, None, None]:
    """Grant home traversal temporarily and restore its original ACL after testing."""
    home = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}"
    root = f"{home}/.lite-nas-acl-{uuid4().hex}"
    path = f"{root}/document.txt"
    backup = f"{home}/.lite-nas-home-acl-{uuid4().hex}"
    operator_cli_client.execute(f"getfacl -p -- {shlex.quote(home)} > {shlex.quote(backup)}")
    operator_cli_client.assert_exit_code(0)
    try:
        operator_cli_client.execute(
            f"setfacl -m {shlex.quote('u:' + TEST_SUDO_LOGIN + ':--x')} -- {shlex.quote(home)} && "
            f"mkdir -m 0755 -- {shlex.quote(root)} && "
            f"printf 'acl-fixture\\n' > {shlex.quote(path)} && "
            f"chmod 0644 -- {shlex.quote(path)}"
        )
        operator_cli_client.assert_exit_code(0)
        yield path
    finally:
        operator_cli_client.execute(f"rm -rf -- {shlex.quote(root)}")
        operator_cli_client.execute(f"setfacl --restore={shlex.quote(backup)}")
        operator_cli_client.assert_exit_code(0)
        operator_cli_client.execute(f"rm -f -- {shlex.quote(backup)}")


# Requirements: file-service-svc/FR-003 and FR-008.
def test_named_user_acl_denial_blocks_file_read(
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    testsudo_cookie: str,
    acl_file: str,
) -> None:
    """Test case: a named-user ACL denial overrides world-readable mode.

    Preparation:
    - The second user can traverse the owner's home to a world-readable file.
    - A named-user ACL entry removes that user's file read permission.

    Action:
    - The denied user requests the file through /f.

    Expected result:
    - RBAC and File Service deny the read.
    """
    operator_cli_client.execute(
        f"setfacl -m {shlex.quote('u:' + TEST_SUDO_LOGIN + ':---')} -- {shlex.quote(acl_file)}"
    )
    operator_cli_client.assert_exit_code(0)
    response = api_client.get(path=f"/f{acl_file}", cookies={ACCESS_COOKIE: testsudo_cookie})
    expect(response.status).to_be(403)


# Requirements: file-service-svc/FR-003.
def test_named_user_acl_grant_allows_file_read(
    api_client: RESTClient,
    operator_cli_client: CLIClient,
    testsudo_cookie: str,
    acl_file: str,
) -> None:
    """Test case: a named-user ACL can grant a private file read access.

    Preparation:
    - The second user can traverse the owner's home to the fixture file.
    - The file is private except for a named read ACL for that user.

    Action:
    - The granted user requests the file through /f.

    Expected result:
    - The service returns the file content.
    """
    operator_cli_client.execute(
        f"chmod 0600 -- {shlex.quote(acl_file)} && "
        f"setfacl -m {shlex.quote('u:' + TEST_SUDO_LOGIN + ':r--')} -- {shlex.quote(acl_file)}"
    )
    operator_cli_client.assert_exit_code(0)
    response = api_client.get(path=f"/f{acl_file}", cookies={ACCESS_COOKIE: testsudo_cookie})
    expect(response.raw_body).to_be("acl-fixture\n")
