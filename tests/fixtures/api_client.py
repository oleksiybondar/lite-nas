import pytest
from constants import API_BASE_URL
from hyperiontf import RESTClient


@pytest.fixture
def api_client() -> RESTClient:
    """Create a REST client for browser-facing LiteNAS API system tests.

    The client accepts HTTP error responses so negative API test cases can make
    the returned status code their single verification point.
    """
    return RESTClient(url=API_BASE_URL, accept_errors=True)
