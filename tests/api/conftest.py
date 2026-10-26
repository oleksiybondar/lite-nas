"""Shared REST fixtures for installed API service tests."""

import os

import pytest
from api.file_service.conftest import login_cookie
from constants import (
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
    SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
)
from fixtures.api_client import SESSION_COOKIE_FILTER
from hyperiontf import RESTClient

MOMENTS_API_URL = os.environ.get("LITENAS_MOMENTS_API_URL", "https://moments.localhost")


@pytest.fixture
def moments_client() -> RESTClient:
    """Connect to the installed standalone Moments host with redacted session logs."""
    client = RESTClient(url=MOMENTS_API_URL, accept_errors=True)
    if SESSION_COOKIE_FILTER not in client.logger.filters:
        client.logger.addFilter(SESSION_COOKIE_FILTER)
    return client


@pytest.fixture
def moments_cookie(moments_client: RESTClient) -> str:
    """Authenticate the operator on the Moments host for metadata and file checks."""
    return login_cookie(
        moments_client,
        SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
        SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
    )
