import logging
import re

import pytest
from constants import API_BASE_URL
from hyperiontf import RESTClient

SESSION_COOKIE_VALUES = re.compile(r"(?i)(lite-nas-(?:at|rt)=)[^;\s,]+")
REFRESH_VALUE_HEADER = re.compile(r"(?im)^(\s*Value:)\s*[^\n]+$")
MAX_LOG_CHARS = 8192


class RedactSessionCookies(logging.Filter):
    """Redact session credentials and bound transfer logs in HyperionTF reports."""

    def filter(self, record: logging.LogRecord) -> bool:
        """Redact cookies and bound large transfer bodies in the test report."""
        message = SESSION_COOKIE_VALUES.sub(r"\1[REDACTED]", record.getMessage())
        message = REFRESH_VALUE_HEADER.sub(r"\1 [REDACTED]", message)
        if len(message) > MAX_LOG_CHARS:
            omitted = len(message) - MAX_LOG_CHARS
            message = (
                f"{message[: MAX_LOG_CHARS // 2]}\n"
                f"[{omitted} transfer-log characters omitted]\n"
                f"{message[-MAX_LOG_CHARS // 2 :]}"
            )
        record.msg = message
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
