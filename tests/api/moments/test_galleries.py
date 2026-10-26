"""Installed-system Moments search and gallery lifecycle checks."""

from http import HTTPStatus
from urllib.parse import urlencode
from uuid import uuid4

import pytest
from api.file_service.conftest import ACCESS_COOKIE
from hyperiontf import RESTClient, expect

from .conftest import must_create_gallery

pytestmark = [pytest.mark.api, pytest.mark.Moments]


# Requirements: moments-backend/FR-001.
@pytest.mark.parametrize("kind", ["ALBUM", "COLLECTION"])
def test_create_and_read_gallery(
    moments_client: RESTClient,
    moments_cookie: str,
    moments_source: tuple[str, str],
    gallery_ids: list[str],
    kind: str,
) -> None:
    """Test case: an owner can read a newly created album or collection.

    Preparation:
    - The operator owns a source directory and creates the selected gallery type.

    Action:
    - Read the new gallery through the Moments API.

    Expected result:
    - Its type is the one chosen at creation.
    """
    directory = moments_source[0] if kind == "ALBUM" else None
    gallery_id = must_create_gallery(
        moments_client, moments_cookie, gallery_ids, kind, f"test-{uuid4().hex}", directory
    )
    response = moments_client.get(
        path=f"/api/galleries/{gallery_id}", cookies={ACCESS_COOKIE: moments_cookie}
    )
    expect(response.body["type"]).to_be(kind)


# Requirements: moments-backend/FR-001 and FR-005.
@pytest.mark.parametrize("kind", ["ALBUM", "COLLECTION"])
def test_search_finds_gallery_by_name(
    moments_client: RESTClient,
    moments_cookie: str,
    moments_source: tuple[str, str],
    gallery_ids: list[str],
    kind: str,
) -> None:
    """Test case: search finds an owned gallery of the requested type.

    Preparation:
    - The operator creates a gallery with a unique name.

    Action:
    - Search the gallery list by name and type.

    Expected result:
    - The created gallery ID appears in the result.
    """
    name = f"search-{uuid4().hex}"
    directory = moments_source[0] if kind == "ALBUM" else None
    gallery_id = must_create_gallery(
        moments_client, moments_cookie, gallery_ids, kind, name, directory
    )
    query = urlencode({"q": name, "type": kind})
    response = moments_client.get(
        path=f"/api/galleries?{query}", cookies={ACCESS_COOKIE: moments_cookie}
    )
    expect(gallery_id in {item["id"] for item in response.body["items"]}).to_be(True)


# Requirements: moments-backend/FR-001.
@pytest.mark.parametrize("kind", ["ALBUM", "COLLECTION"])
def test_owner_can_rename_gallery(
    moments_client: RESTClient,
    moments_cookie: str,
    moments_source: tuple[str, str],
    gallery_ids: list[str],
    kind: str,
) -> None:
    """Test case: an owner can rename an album or collection.

    Preparation:
    - The operator creates the selected gallery type.

    Action:
    - PATCH its name.

    Expected result:
    - The response contains the new name.
    """
    directory = moments_source[0] if kind == "ALBUM" else None
    gallery_id = must_create_gallery(
        moments_client, moments_cookie, gallery_ids, kind, f"old-{uuid4().hex}", directory
    )
    new_name = f"renamed-{uuid4().hex}"
    response = moments_client.patch(
        path=f"/api/galleries/{gallery_id}",
        payload={"name": new_name},
        cookies={ACCESS_COOKIE: moments_cookie},
    )
    expect(response.body["name"]).to_be(new_name)


# Requirements: moments-backend/FR-001.
@pytest.mark.parametrize("kind", ["ALBUM", "COLLECTION"])
def test_owner_can_delete_gallery(
    moments_client: RESTClient,
    moments_cookie: str,
    moments_source: tuple[str, str],
    gallery_ids: list[str],
    kind: str,
) -> None:
    """Test case: deleting gallery metadata makes it unavailable by ID.

    Preparation:
    - The operator creates the selected gallery type.

    Action:
    - DELETE it and then request its detail.

    Expected result:
    - The detail route returns not found.
    """
    directory = moments_source[0] if kind == "ALBUM" else None
    gallery_id = must_create_gallery(
        moments_client, moments_cookie, gallery_ids, kind, f"delete-{uuid4().hex}", directory
    )
    deleted = moments_client.delete(
        path=f"/api/galleries/{gallery_id}", cookies={ACCESS_COOKIE: moments_cookie}
    )
    if deleted.status != HTTPStatus.NO_CONTENT:
        pytest.fail(f"gallery deletion failed: HTTP {deleted.status}")
    response = moments_client.get(
        path=f"/api/galleries/{gallery_id}", cookies={ACCESS_COOKIE: moments_cookie}
    )
    expect(response.status).to_be(HTTPStatus.NOT_FOUND)
