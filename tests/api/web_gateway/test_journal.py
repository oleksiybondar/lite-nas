"""Installed-system journal gateway checks using HyperionTF RESTClient."""

from http import HTTPStatus

import pytest
from hyperiontf import RESTClient, expect

pytestmark = [pytest.mark.api, pytest.mark.WebGateway, pytest.mark.JournalManager]


def test_security_role_can_query_journal(api_client: RESTClient, security_cookie: str) -> None:
    """Test case: a security user can browse a bounded journal chunk.

    Preparation:
    - The journal manager and gateway are installed and running.
    - The security test user is authenticated.

    Action:
    - Request ten entries from the host journal.

    Expected result:
    - The gateway returns one valid journal chunk.
    """
    response = api_client.get(
        path="/api/journal?limit=10", cookies={"lite-nas-at": security_cookie}
    )
    if response.status != HTTPStatus.OK:
        pytest.fail(f"journal query failed: HTTP {response.status}")
    expect(isinstance(response.body["data"]["entries"], list)).to_be(True)


def test_security_role_can_list_journal_units(api_client: RESTClient, security_cookie: str) -> None:
    """Test case: a security user can list service units found in the journal.

    Preparation:
    - The journal manager and gateway are installed and running.
    - The security test user is authenticated.

    Action:
    - Request journal units without filters.

    Expected result:
    - The gateway returns a list of matching unit names.
    """
    response = api_client.get(path="/api/journal/units", cookies={"lite-nas-at": security_cookie})
    if response.status != HTTPStatus.OK:
        pytest.fail(f"journal unit query failed: HTTP {response.status}")
    expect(isinstance(response.body["data"]["units"], list)).to_be(True)


@pytest.mark.parametrize("query", ["unit=--help", "limit=101", "cursor=bad%20cursor"])
def test_journal_rejects_invalid_query(
    api_client: RESTClient, security_cookie: str, query: str
) -> None:
    """Test case: malformed journal filters never reach command execution.

    Preparation:
    - The installed gateway and journal manager are running.
    - The security test user is authenticated.

    Action:
    - Request one malformed journal filter.

    Expected result:
    - Gateway validation rejects it.
    """
    response = api_client.get(
        path=f"/api/journal?{query}", cookies={"lite-nas-at": security_cookie}
    )
    expect(response.status).to_be(HTTPStatus.UNPROCESSABLE_ENTITY)


def test_vacuum_rejects_invalid_days(api_client: RESTClient, security_cookie: str) -> None:
    """Test case: invalid retention cannot invoke the privileged vacuum command.

    Preparation:
    - The installed gateway and journal manager are running.
    - The security test user is authenticated.

    Action:
    - Submit a zero-day retention request.

    Expected result:
    - Gateway validation rejects the request before vacuuming.
    """
    response = api_client.post(
        path="/api/journal/vacuum", payload={"days": 0}, cookies={"lite-nas-at": security_cookie}
    )
    expect(response.status).to_be(HTTPStatus.UNPROCESSABLE_ENTITY)
