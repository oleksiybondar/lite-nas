"""Installed-system Moments search and source-path injection checks."""

import shlex
from http import HTTPStatus
from urllib.parse import urlencode
from uuid import uuid4

import pytest
from api.file_service.conftest import ACCESS_COOKIE, injection_name
from constants import SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN
from hyperiontf import CLIClient, RESTClient, expect

from .conftest import must_create_gallery

pytestmark = [pytest.mark.api, pytest.mark.Moments]


# Requirements: moments-backend/FR-005.
def test_sql_syntax_in_search_is_literal(
    moments_client: RESTClient, moments_cookie: str, gallery_ids: list[str]
) -> None:
    """Test case: SQL syntax in search cannot expand gallery results.

    Preparation:
    - The operator owns a uniquely named collection.

    Action:
    - Search for a SQL tautology string.

    Expected result:
    - The owned collection is absent from the result.
    """
    gallery_id = must_create_gallery(
        moments_client, moments_cookie, gallery_ids, "COLLECTION", f"literal-{uuid4().hex}"
    )
    query = urlencode({"q": "' OR 1=1 --"})
    response = moments_client.get(
        path=f"/api/galleries?{query}",
        cookies={ACCESS_COOKIE: moments_cookie},
    )
    expect(gallery_id not in {item["id"] for item in response.body["items"]}).to_be(True)


# Requirements: moments-backend/FR-001 and FR-003.
def test_shell_syntax_in_album_path_does_not_execute(
    moments_client: RESTClient,
    moments_cookie: str,
    operator_cli_client: CLIClient,
    moments_source: tuple[str, str],
    gallery_ids: list[str],
) -> None:
    """Test case: a source directory name cannot execute embedded shell syntax.

    Preparation:
    - The operator owns a disposable directory and an absent execution sentinel.

    Action:
    - Create an album using a child path with shell syntax.

    Expected result:
    - The sentinel remains absent after path validation.
    """
    directory, _ = moments_source
    sentinel = (
        f"/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}/"
        f".lite-nas-moments-{uuid4().hex}.sentinel"
    )
    source = f"{directory}/{injection_name(sentinel)}"
    response = moments_client.post(
        path="/api/galleries",
        payload={"type": "ALBUM", "name": f"injection-{uuid4().hex}", "uploadDirectory": source},
        cookies={ACCESS_COOKIE: moments_cookie},
    )
    if response.status == HTTPStatus.OK:
        gallery_ids.append(str(response.body["id"]))
    elif response.status not in (HTTPStatus.NOT_FOUND, HTTPStatus.UNPROCESSABLE_ENTITY):
        pytest.fail(f"unexpected path validation result: HTTP {response.status}")
    operator_cli_client.execute(f"test ! -e {shlex.quote(sentinel)}")
    operator_cli_client.assert_exit_code(0)
