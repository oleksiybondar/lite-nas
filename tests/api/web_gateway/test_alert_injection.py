"""Installed-system SQL filter checks for browser-facing alerts."""

import json
from http import HTTPStatus
from urllib.parse import quote

import pytest
from constants import (
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
    SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
)
from hyperiontf import RESTClient, expect

pytestmark = [pytest.mark.api, pytest.mark.WebGateway, pytest.mark.SystemLoggingManager]


def test_sql_syntax_in_alert_filter_is_a_literal_value(api_client: RESTClient) -> None:
    """Test case: a SQL-shaped category cannot broaden alert search results.

    Preparation:
    - The installed operator logs in and can request system alerts.

    Action:
    - Query alerts with a category filter containing a SQL tautology.

    Expected result:
    - The gateway and logging manager return no matching alerts.
    """
    login = api_client.post(
        path="/api/auth/login",
        payload={
            "login": SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
            "password": SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
        },
    )
    if login.status != HTTPStatus.OK or "lite-nas-at" not in login.cookies:
        pytest.fail(f"operator login failed: HTTP {login.status}")
    value = json.dumps({"key": "category", "condition": "eq", "values": ["' OR 1=1 --"]})
    response = api_client.get(
        path=f"/api/alerts/system?filters={quote(value, safe='')}",
        cookies={"lite-nas-at": str(login.cookies["lite-nas-at"])},
    )
    if response.status != HTTPStatus.OK:
        pytest.fail(f"alert filter request failed: HTTP {response.status}")
    expect(response.body["data"]["metadata"]["total_count"]).to_be(0)
