"""Installed REST service response contracts validated with HyperionTF."""

from http import HTTPStatus

import pytest
from api.file_service.conftest import ACCESS_COOKIE, login_cookie
from constants import (
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
    SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
)
from hyperiontf import RESTClient
from schemas.rest_contracts import (
    AUTH_ME_SCHEMA,
    FILE_STAT_SCHEMA,
    GALLERY_LIST_SCHEMA,
    STREAM_OPENAPI_SCHEMA,
)

pytestmark = [pytest.mark.api, pytest.mark.contract]


@pytest.fixture
def contract_cookie(api_client: RESTClient) -> str:
    """Log in the installed operator for gateway and File Service contracts."""
    return login_cookie(
        api_client,
        SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
        SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
    )


# Requirements: moments/FR-001.
@pytest.mark.WebGateway
@pytest.mark.Auth
def test_gateway_current_user_response_schema(api_client: RESTClient, contract_cookie: str) -> None:
    """Test case: the gateway current-user response matches its JSON contract.

    Preparation:
    - The installed operator has an authenticated browser session.

    Action:
    - GET the gateway current-user endpoint.

    Expected result:
    - Its JSON response satisfies the documented session shape.
    """
    response = api_client.get(path="/api/auth/me", cookies={ACCESS_COOKIE: contract_cookie})
    if response.status != HTTPStatus.OK:
        pytest.fail(f"current-user lookup failed: HTTP {response.status}")
    response.validate_json_schema(AUTH_ME_SCHEMA)


# Requirements: file-service-svc/FR-008.
@pytest.mark.FileService
def test_file_stat_response_schema(api_client: RESTClient, contract_cookie: str) -> None:
    """Test case: File Service stat returns its published JSON fields.

    Preparation:
    - The installed operator can read its own home directory.

    Action:
    - GET stat for that directory through the public file mount.

    Expected result:
    - The response matches the file stat JSON Schema.
    """
    path = f"/f/stat/home/{SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN}"
    response = api_client.get(path=path, cookies={ACCESS_COOKIE: contract_cookie})
    if response.status != HTTPStatus.OK:
        pytest.fail(f"file stat failed: HTTP {response.status}")
    response.validate_json_schema(FILE_STAT_SCHEMA)


# Requirements: moments-backend/FR-001 and FR-005.
@pytest.mark.Moments
def test_moments_gallery_list_response_schema(
    moments_client: RESTClient, moments_cookie: str
) -> None:
    """Test case: Moments gallery search returns its paginated JSON contract.

    Preparation:
    - The installed operator has a Moments session.

    Action:
    - GET the accessible gallery list.

    Expected result:
    - Pagination and every returned gallery satisfy the JSON Schema.
    """
    response = moments_client.get(path="/api/galleries", cookies={ACCESS_COOKIE: moments_cookie})
    if response.status != HTTPStatus.OK:
        pytest.fail(f"Moments gallery list failed: HTTP {response.status}")
    response.validate_json_schema(GALLERY_LIST_SCHEMA)


# Requirements: stream-service-svc/FR-001 and IR-001.
@pytest.mark.StreamService
def test_stream_openapi_response_schema(api_client: RESTClient) -> None:
    """Test case: Stream Service publishes the JSON contract for its media routes.

    Preparation:
    - Stream Service is installed through the public Nginx mount.

    Action:
    - GET its OpenAPI JSON document.

    Expected result:
    - The document names the original, preview, and transcode routes.
    """
    response = api_client.get(path="/s/openapi.json")
    if response.status != HTTPStatus.OK:
        pytest.fail(f"Stream Service OpenAPI lookup failed: HTTP {response.status}")
    response.validate_json_schema(STREAM_OPENAPI_SCHEMA)
