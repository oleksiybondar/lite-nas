"""System API test suite for browser-facing authentication behavior."""

import pytest
from constants import CREDENTIALS
from hyperiontf import RESTClient, expect

LOGIN_ENDPOINT = "/api/auth/login"


@pytest.mark.Auth
@pytest.mark.WebGateway
@pytest.mark.api
def test_login_returns_authenticated_session(api_client: RESTClient) -> None:
    """Test case: valid login returns an authenticated session.

    Preparation:
    - The LiteNAS auth service and web gateway are running.
    - The configured test user exists with the configured password.

    Action:
    - Submit the configured credentials to the browser-facing login endpoint.

    Expected result:
    - The response reports that the session is authenticated.
    """
    response = api_client.post(
        path=LOGIN_ENDPOINT,
        payload={"login": CREDENTIALS["login"], "password": CREDENTIALS["password"]},
    )
    expect(response.body["data"]["authenticated"]).to_be(True)


@pytest.mark.Auth
@pytest.mark.WebGateway
@pytest.mark.api
def test_login_rejects_invalid_credentials(api_client: RESTClient) -> None:
    """Test case: invalid login credentials are rejected.

    Preparation:
    - The LiteNAS auth service and web gateway are running.

    Action:
    - Submit credentials that do not belong to a valid authenticated user.

    Expected result:
    - The login endpoint returns an unauthorized response.
    """
    response = api_client.post(
        path=LOGIN_ENDPOINT,
        payload={"login": "invalid-user", "password": "wrong-password"},
    )
    expect(response.status).to_be(401)
