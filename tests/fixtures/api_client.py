import logging
import re

import pytest
from constants import API_BASE_URL
from hyperiontf import RESTClient

SESSION_COOKIE_VALUES = re.compile(r"(?i)(lite-nas-(?:at|rt)=)[^;\s,]+")
REFRESH_VALUE_HEADER = re.compile(r"(?im)^(\s*Value:)\s*[^\n]+$")


class RedactSessionCookies(logging.Filter):
    """Remove session credentials from HyperionTF request and response logs."""

    def filter(self, record: logging.LogRecord) -> bool:
        """Redact access and refresh values before a test report records them."""
        message = SESSION_COOKIE_VALUES.sub(r"\1[REDACTED]", record.getMessage())
        record.msg = REFRESH_VALUE_HEADER.sub(r"\1 [REDACTED]", message)
        record.args = ()
        return True


SESSION_COOKIE_FILTER = RedactSessionCookies()


@pytest.fixture
def api_client() -> RESTClient:
    """Create a REST client for browser-facing LiteNAS API system tests.

    The client accepts HTTP error responses so negative API test cases can make
    the returned status code their single verification point.
    """
    client = RESTClient(url=API_BASE_URL, accept_errors=True)
    if SESSION_COOKIE_FILTER not in client.logger.filters:
        client.logger.addFilter(SESSION_COOKIE_FILTER)
    return client
