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


@pytest.mark.Auth
@pytest.mark.WebGateway
@pytest.mark.api
@pytest.mark.parametrize(
    ("login", "status"),
    [
        pytest.param("testuser' OR '1'='1", 422, id="sql-shaped-login"),
        pytest.param("testuser\nroot", 422, id="newline-login"),
        pytest.param("--help", 401, id="option-shaped-pam-username"),
    ],
)
def test_login_injection_strings_do_not_authenticate(
    api_client: RESTClient, login: str, status: int
) -> None:
    """Test case: login syntax cannot change the principal or PAM invocation.

    Preparation:
    - The configured test account exists with the configured password.

    Action:
    - Submit the selected malformed or option-shaped login with that password.

    Expected result:
    - Web Gateway validation or PAM denies the login without a session.
    """
    response = api_client.post(
        path=LOGIN_ENDPOINT,
        payload={"login": login, "password": CREDENTIALS["password"]},
    )
    expect(response.status).to_be(status)


@pytest.mark.Auth
@pytest.mark.WebGateway
@pytest.mark.api
def test_login_password_sql_syntax_is_not_an_authentication_bypass(
    api_client: RESTClient,
) -> None:
    """Test case: SQL-shaped password text is treated as an incorrect PAM secret.

    Preparation:
    - The configured test account exists with a different password.

    Action:
    - Submit its real login with SQL tautology text as the password.

    Expected result:
    - Authentication returns 401 without issuing a session.
    """
    response = api_client.post(
        path=LOGIN_ENDPOINT,
        payload={"login": CREDENTIALS["login"], "password": "' OR '1'='1"},
    )
    expect(response.status).to_be(401)
