"""Installed-system Moments collection membership checks."""

from http import HTTPStatus
from uuid import uuid4

import pytest
from api.file_service.conftest import ACCESS_COOKIE
from hyperiontf import RESTClient, expect

from .conftest import must_create_gallery

pytestmark = [pytest.mark.api, pytest.mark.Moments]


# Requirements: moments-backend/FR-002 and FR-003.
@pytest.mark.parametrize(
    ("kind", "source_index"),
    [("files", 1), ("directories", 0)],
)
def test_add_collection_reference(  # noqa: PLR0913
    moments_client: RESTClient,
    moments_cookie: str,
    moments_source: tuple[str, str],
    gallery_ids: list[str],
    kind: str,
    source_index: int,
) -> None:
    """Test case: an owner can add an authorized source to a collection.

    Preparation:
    - The operator owns a file and directory and creates a collection.

    Action:
    - Add the selected source and list collection references.

    Expected result:
    - The source path appears in membership.
    """
    gallery_id = must_create_gallery(
        moments_client, moments_cookie, gallery_ids, "COLLECTION", f"refs-{uuid4().hex}"
    )
    path = moments_source[source_index]
    added = moments_client.post(
        path=f"/api/collections/{gallery_id}/{kind}",
        payload={"path": path},
        cookies={ACCESS_COOKIE: moments_cookie},
    )
    if added.status != HTTPStatus.OK:
        pytest.fail(f"reference addition failed: HTTP {added.status}")
    response = moments_client.get(
        path=f"/api/collections/{gallery_id}/references",
        cookies={ACCESS_COOKIE: moments_cookie},
    )
    expect(path in {item["path"] for item in response.body["items"]}).to_be(True)


# Requirements: moments-backend/FR-002.
@pytest.mark.parametrize(
    ("kind", "source_index"),
    [("files", 1), ("directories", 0)],
)
def test_remove_collection_reference(  # noqa: PLR0913
    moments_client: RESTClient,
    moments_cookie: str,
    moments_source: tuple[str, str],
    gallery_ids: list[str],
    kind: str,
    source_index: int,
) -> None:
    """Test case: removing membership leaves the collection without that source.

    Preparation:
    - The operator adds an owned source to a disposable collection.

    Action:
    - Remove the selected source and list references.

    Expected result:
    - The source path is absent from membership.
    """
    gallery_id = must_create_gallery(
        moments_client, moments_cookie, gallery_ids, "COLLECTION", f"remove-{uuid4().hex}"
    )
    path = moments_source[source_index]
    added = moments_client.post(
        path=f"/api/collections/{gallery_id}/{kind}",
        payload={"path": path},
        cookies={ACCESS_COOKIE: moments_cookie},
    )
    if added.status != HTTPStatus.OK:
        pytest.fail(f"reference setup failed: HTTP {added.status}")
    removed = moments_client.delete(
        path=f"/api/collections/{gallery_id}/{kind}",
        payload={"path": path},
        cookies={ACCESS_COOKIE: moments_cookie},
    )
    if removed.status != HTTPStatus.OK:
        pytest.fail(f"reference removal failed: HTTP {removed.status}")
    response = moments_client.get(
        path=f"/api/collections/{gallery_id}/references",
        cookies={ACCESS_COOKIE: moments_cookie},
    )
    expect(path not in {item["path"] for item in response.body["items"]}).to_be(True)
