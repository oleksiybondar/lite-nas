"""Installed-system stream access boundaries through the Nginx media mount."""

from http import HTTPStatus

import pytest
from api.file_service.conftest import ACCESS_COOKIE
from hyperiontf import RESTClient, expect

pytestmark = [pytest.mark.api, pytest.mark.StreamService, pytest.mark.Nginx]


# Requirements: stream-service-svc/FR-001 and IR-001.
@pytest.mark.parametrize("route", ["/s", "/s/transcode", "/s/preview"])
def test_anonymous_media_route_is_denied(
    api_client: RESTClient, readable_media: str, route: str
) -> None:
    """Test case: an unauthenticated request cannot access original or converted media.

    Preparation:
    - Stream Service and Nginx are installed, and a readable source exists.

    Action:
    - Request the source through the selected route without a token.

    Expected result:
    - Authentication rejects the request before media bytes are returned.
    """
    response = api_client.get(path=f"{route}{readable_media}")
    expect(response.status).to_be(HTTPStatus.UNAUTHORIZED)


# Requirements: stream-service-svc/FR-003 and FR-004.
def test_owner_can_read_original_range(
    api_client: RESTClient, operator_cookie: str, readable_media: str
) -> None:
    """Test case: the verified owner receives a partial original-file response.

    Preparation:
    - The operator owns an AVI whose ACL grants the service process read access.

    Action:
    - Request the first byte with the operator's access cookie.

    Expected result:
    - The service returns HTTP 206 after the caller-specific RBAC read decision.
    """
    response = api_client.get(
        path=f"/s{readable_media}",
        headers={"Range": "bytes=0-0"},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(response.status).to_be(HTTPStatus.PARTIAL_CONTENT)


# Requirements: stream-service-svc/FR-003, FR-004, and SR-001.
def test_owner_can_stream_private_file_through_user_reader(
    api_client: RESTClient, operator_cookie: str, operator_owned_file: str
) -> None:
    """Test case: a private file uses its verified owner's Unix read rights.

    Preparation:
    - The operator owns a 0600 file created after service ACL provisioning.

    Action:
    - Request a one-byte range with the operator's access cookie.

    Expected result:
    - The per-user reader returns HTTP 206 without a root-level file exception.
    """
    response = api_client.get(
        path=f"/s{operator_owned_file}",
        headers={"Range": "bytes=0-0"},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(response.status).to_be(HTTPStatus.PARTIAL_CONTENT)


# Requirements: stream-service-svc/FR-003 and SR-001.
def test_other_user_cannot_read_operator_media(
    api_client: RESTClient, testsudo_cookie: str, readable_media: str
) -> None:
    """Test case: process-level read access does not grant another user media access.

    Preparation:
    - The operator's file grants service read access, and another user logs in.

    Action:
    - The other user requests a one-byte range from the operator's file.

    Expected result:
    - RBAC prevents a media response.
    """
    response = api_client.get(
        path=f"/s{readable_media}",
        headers={"Range": "bytes=0-0"},
        cookies={ACCESS_COOKIE: testsudo_cookie},
    )
    expect(response.status).to_be(HTTPStatus.NOT_FOUND)


# Requirements: stream-service-svc/FR-005 and TR-001.
def test_owner_can_transcode_valid_avi(
    api_client: RESTClient, operator_cookie: str, valid_avi: str
) -> None:
    """Test case: an authorized AVI converts to browser-compatible media.

    Preparation:
    - The operator owns a valid short AVI and has an authenticated session.

    Action:
    - Request the explicit fMP4 fallback URL through the Nginx stream mount.

    Expected result:
    - The service returns HTTP 200 after user-level RBAC and file checks.
    """
    response = api_client.get(
        path=f"/s/transcode{valid_avi}?start=0",
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(response.status).to_be(HTTPStatus.OK)


# Requirements: stream-service-svc/FR-005 and TR-001.
def test_browser_range_header_receives_sequential_transcode(
    api_client: RESTClient, operator_cookie: str, valid_avi: str
) -> None:
    """Test case: a browser Range header does not prevent transformed playback.

    Preparation:
    - The operator owns a valid AVI and has an authenticated session.

    Action:
    - Request the transcoded URL with the Range header used by HTML5 video.

    Expected result:
    - The service returns a sequential HTTP 200 media response.
    """
    response = api_client.get(
        path=f"/s/transcode{valid_avi}?start=0",
        headers={"Range": "bytes=0-"},
        cookies={ACCESS_COOKIE: operator_cookie},
    )
    expect(response.status).to_be(HTTPStatus.OK)


# Requirements: stream-service-svc/FR-003 and SR-001.
def test_other_user_cannot_transcode_operator_media(
    api_client: RESTClient, testsudo_cookie: str, valid_avi: str
) -> None:
    """Test case: a URL to transcode private media grants no access by itself.

    Preparation:
    - The operator owns a private AVI and another user has a valid session.

    Action:
    - The other user requests the operator's transcoded video URL.

    Expected result:
    - RBAC denies the read before ffmpeg starts.
    """
    response = api_client.get(
        path=f"/s/transcode{valid_avi}?start=0",
        cookies={ACCESS_COOKIE: testsudo_cookie},
    )
    expect(response.status).to_be(HTTPStatus.NOT_FOUND)


# Requirements: stream-service-svc/FR-002 and SR-001.
def test_other_user_cannot_request_operator_preview(
    api_client: RESTClient, testsudo_cookie: str, readable_media: str
) -> None:
    """Test case: a JPEG preview URL does not bypass the owner's file ACL.

    Preparation:
    - The operator owns a file and another user has a valid session.

    Action:
    - The other user requests its preview route.

    Expected result:
    - RBAC rejects the path before conversion or format inspection.
    """
    response = api_client.get(
        path=f"/s/preview{readable_media}",
        cookies={ACCESS_COOKIE: testsudo_cookie},
    )
    expect(response.status).to_be(HTTPStatus.NOT_FOUND)


# Requirements: stream-service-svc/SR-001.
@pytest.mark.parametrize(
    "host_path",
    ["/etc/passwd", "/etc/lite-nas/auth.conf", "/var/log/lite-nas/auth-service.log"],
)
def test_system_paths_are_outside_media_scope(
    api_client: RESTClient, operator_cookie: str, host_path: str
) -> None:
    """Test case: an authenticated caller cannot stream a system file.

    Preparation:
    - The selected system path exists on an installed LiteNAS host.

    Action:
    - Request it through the stream mount with a valid session.

    Expected result:
    - Storage scope rejects the path before any file bytes are served.
    """
    response = api_client.get(path=f"/s{host_path}", cookies={ACCESS_COOKIE: operator_cookie})
    expect(response.status).to_be(HTTPStatus.NOT_FOUND)
