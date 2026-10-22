"""Installed-system directory listing checks through the public /ls mount."""

import shlex
from collections.abc import Generator
from pathlib import PurePosixPath
from uuid import uuid4

import pytest
from constants import SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN
from hyperiontf import CLIClient, RESTClient, expect

from .conftest import ACCESS_COOKIE

pytestmark = [pytest.mark.api, pytest.mark.FileService, pytest.mark.Nginx]

SYSTEM_DIRECTORIES = {"/boot", "/dev", "/etc", "/proc", "/run", "/sys", "/usr", "/var"}


# Requirements: file-service-svc/FR-001 and file-service-svc/FR-008.
def test_anonymous_root_listing_is_denied(api_client: RESTClient) -> None:
    """Test case: an anonymous caller cannot enumerate filesystem roots.

    Preparation:
    - File Service is available through Nginx's public /ls mount.

    Action:
    1. GET /ls/ without an access token.

    Expected result:
    - The response is 401 and exposes no root entries.
    """
    response = api_client.get(path="/ls/")
    expect(response.status).to_be(401)


@pytest.fixture
def zfs_root_mounts(cli_client: CLIClient) -> set[str]:
    """Discover top-level directory entries containing mounted ZFS filesystems."""
    cli_client.execute("zfs list -H -o mountpoint -t filesystem")
    cli_client.assert_exit_code(0)
    return {
        f"/{PurePosixPath(mountpoint).parts[1]}"
        for mountpoint in cli_client.output.splitlines()
        if mountpoint.startswith("/") and mountpoint != "/"
    }


# Requirements: file-service-svc/FR-008.
@pytest.mark.parametrize("account", ["operator", "testsudo"])
def test_non_admin_root_listing_contains_home_and_pools(
    api_client: RESTClient,
    request: pytest.FixtureRequest,
    account: str,
    zfs_root_mounts: set[str],
) -> None:
    """Test case: non-admin accounts see home and mounted ZFS pool roots.

    Preparation:
    - The operator or restricted sudo user has a browser session.
    - Mounted ZFS filesystem paths are discovered with zfs list.

    Action:
    1. GET /ls/ with that account's access cookie.

    Expected result:
    - The root listing contains /home and each mounted ZFS top-level directory.
    """
    cookie = request.getfixturevalue(f"{account}_cookie")
    response = api_client.get(path="/ls/", cookies={ACCESS_COOKIE: cookie})
    expect(response.status).to_be(200)
    expect({"/home", *zfs_root_mounts} <= set(response.body["directories"])).to_be(True)


# Requirements: file-service-svc/FR-008.
@pytest.mark.parametrize("account", ["operator", "testsudo"])
def test_non_admin_root_listing_omits_system_directories(
    api_client: RESTClient, request: pytest.FixtureRequest, account: str
) -> None:
    """Test case: non-admin accounts cannot enumerate system directory names.

    Preparation:
    - The operator or restricted sudo user has a browser session.

    Action:
    1. GET /ls/ with that account's access cookie.

    Expected result:
    - System directories including /etc, /dev, /usr, /var, /sys, and /proc
      are absent from the root listing.
    """
    cookie = request.getfixturevalue(f"{account}_cookie")
    response = api_client.get(path="/ls/", cookies={ACCESS_COOKIE: cookie})
    expect(response.status).to_be(200)
    expect(SYSTEM_DIRECTORIES.isdisjoint(response.body["directories"])).to_be(True)


# Requirements: file-service-svc/FR-008.
def test_admin_root_listing_contains_system_directories(
    api_client: RESTClient, admin_cookie: str
) -> None:
    """Test case: admin can enumerate readable system directories at root.

    Preparation:
    - The configured File Service admin has a browser session.

    Action:
    1. GET /ls/ with the admin access cookie.

    Expected result:
    - Standard system directories including /etc, /var, /sys, and /proc are
      included in the root listing.
    """
    response = api_client.get(path="/ls/", cookies={ACCESS_COOKIE: admin_cookie})
    expect(response.status).to_be(200)
    expect(SYSTEM_DIRECTORIES.issubset(response.body["directories"])).to_be(True)


@pytest.fixture
def listing_tree(
    operator_cli_client: CLIClient,
) -> Generator[tuple[str, str, str], None, None]:
    """Create a private two-level listing tree and remove it after the test."""
    root = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-ls-{uuid4().hex}"
    nested = f"{root}/nested"
    direct_file = f"{root}/direct.txt"
    nested_file = f"{nested}/nested.txt"
    operator_cli_client.execute(
        f"mkdir -m 0700 -p -- {shlex.quote(nested)} && "
        f"printf direct > {shlex.quote(direct_file)} && "
        f"printf nested > {shlex.quote(nested_file)}"
    )
    operator_cli_client.assert_exit_code(0)
    try:
        yield root, direct_file, nested_file
    finally:
        operator_cli_client.execute(f"rm -rf -- {shlex.quote(root)}")


# Requirements: file-service-svc/FR-008.
def test_listing_returns_direct_children_only(
    api_client: RESTClient, operator_cookie: str, listing_tree: tuple[str, str, str]
) -> None:
    """Test case: a plain listing stays within the requested directory.

    Preparation:
    - The operator owns a directory containing a file and a nested directory.

    Action:
    - List the directory without a recursive flag.

    Expected result:
    - The response contains only the direct file and directory entries.
    """
    root, direct_file, _ = listing_tree
    response = api_client.get(path=f"/ls{root}", cookies={ACCESS_COOKIE: operator_cookie})
    expect(response.body).to_be({"files": [direct_file], "directories": [f"{root}/nested"]})


# Requirements: file-service-svc/FR-008.
def test_recursive_listing_returns_entire_tree(
    api_client: RESTClient, operator_cookie: str, listing_tree: tuple[str, str, str]
) -> None:
    """Test case: recursive listing returns the complete authorized tree.

    Preparation:
    - The operator owns a directory with a file in an accessible child directory.

    Action:
    1. GET the parent through /ls with the recursive flag.

    Expected result:
    - The response contains both files and the nested directory as absolute paths.
    """
    root, direct_file, nested_file = listing_tree
    response = api_client.get(path=f"/ls{root}?recursive", cookies={ACCESS_COOKIE: operator_cookie})
    expect(response.body).to_be(
        {"files": [direct_file, nested_file], "directories": [f"{root}/nested"]}
    )


# Requirements: file-service-svc/FR-008.
def test_files_only_lists_direct_files(
    api_client: RESTClient, operator_cookie: str, listing_tree: tuple[str, str, str]
) -> None:
    """Test case: files-only lists direct files without descending.

    Preparation:
    - The operator owns a directory with one direct file and one nested file.

    Action:
    1. GET the parent through /ls with only the files-only flag.

    Expected result:
    - Only the direct file appears and the directories array is empty.
    """
    root, direct_file, _ = listing_tree
    response = api_client.get(
        path=f"/ls{root}?files-only", cookies={ACCESS_COOKIE: operator_cookie}
    )
    expect(response.body).to_be({"files": [direct_file], "directories": []})


# Requirements: file-service-svc/FR-008.
def test_recursive_files_only_lists_all_files(
    api_client: RESTClient, operator_cookie: str, listing_tree: tuple[str, str, str]
) -> None:
    """Test case: combined flags return files at every accessible depth.

    Preparation:
    - The operator owns one direct file and one file in an accessible child directory.

    Action:
    1. GET the parent through /ls with recursive and files-only flags present.

    Expected result:
    - Both files appear as absolute paths and directories is empty.
    """
    root, direct_file, nested_file = listing_tree
    response = api_client.get(
        path=f"/ls{root}?recursive&files-only",
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(response.body).to_be({"files": [direct_file, nested_file], "directories": []})


# Requirements: file-service-svc/FR-008.
def test_missing_directory_returns_error(api_client: RESTClient, operator_cookie: str) -> None:
    """Test case: a missing directory returns a documented error.

    Preparation:
    - The operator has a valid browser session.

    Action:
    - List a unique absent path beneath the operator's home.

    Expected result:
    - The service returns a JSON not-found error.
    """
    path = f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/.lite-nas-missing-{uuid4().hex}"
    response = api_client.get(path=f"/ls{path}", cookies={ACCESS_COOKIE: operator_cookie})
    expect(response.body["error"]["code"]).to_be("not_found")
