"""Installed Moments picker configuration contract checks."""

from http import HTTPStatus

import pytest
from api.file_service.conftest import ACCESS_COOKIE, login_cookie
from constants import (
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
    SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
)
from hyperiontf import RESTClient, expect
from schemas.rest_contracts import MOMENTS_CONFIG_SCHEMA

pytestmark = [pytest.mark.api, pytest.mark.Moments]


# Requirements: moments/FR-004.
def test_config_returns_working_directory(moments_client: RESTClient, moments_cookie: str) -> None:
    """Test case: signed-in Moments users receive the picker start path.

    Preparation:
    - The installed service has a working_dir setting and an operator session.

    Action:
    - Request the authenticated Moments configuration endpoint.

    Expected result:
    - The response contains one absolute workingDir path.
    """
    response = moments_client.get(path="/api/config", cookies={ACCESS_COOKIE: moments_cookie})
    if response.status != HTTPStatus.OK:
        pytest.fail(f"Moments configuration failed: HTTP {response.status}")
    response.validate_json_schema(MOMENTS_CONFIG_SCHEMA)


# Requirements: moments/FR-004.
def test_config_rejects_anonymous_request(moments_client: RESTClient) -> None:
    """Test case: an anonymous caller cannot read the picker start path.

    Preparation:
    - The installed Moments API is available without a session cookie.

    Action:
    - Request the Moments configuration endpoint.

    Expected result:
    - The API rejects the request as unauthorized.
    """
    response = moments_client.get(path="/api/config")
    expect(response.status).to_be(HTTPStatus.UNAUTHORIZED)


# Requirements: moments/FR-004.
def test_embedded_moments_config_uses_its_own_api(api_client: RESTClient) -> None:
    """Test case: embedded Moments can read its configuration through the admin host.

    Preparation:
    - The installed operator signs in on the admin host.

    Action:
    - Request the embedded Moments API mount's configuration route.

    Expected result:
    - Nginx forwards the request to Moments and returns its configuration.
    """
    cookie = login_cookie(
        api_client,
        SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
        SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
    )
    response = api_client.get(path="/apps/moments/api/config", cookies={ACCESS_COOKIE: cookie})
    response.validate_json_schema(MOMENTS_CONFIG_SCHEMA)
