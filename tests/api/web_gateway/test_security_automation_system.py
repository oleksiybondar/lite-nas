"""Installed gateway tests for Security Automation authorization and events."""

import time
from http import HTTPStatus

import pytest
from api.file_service.conftest import ACCESS_COOKIE
from constants import SECURITY_LOGGING_MANAGER_LOGIN, SECURITY_LOGGING_MANAGER_PASSWORD
from fixtures.security_automation import run_automation_cli
from hyperiontf import CLIClient, RESTClient, expect

pytestmark = [pytest.mark.api, pytest.mark.SecurityAutomation, pytest.mark.WebGateway]
BASE = "/api/security/automations"


@pytest.mark.parametrize("identity", ["operator_cookie", "testsudo_cookie"])
@pytest.mark.parametrize("path", [f"{BASE}/rules", f"{BASE}/active"])
def test_operator_cannot_read_automation_api(
    api_client: RESTClient,
    request: pytest.FixtureRequest,
    identity: str,
    path: str,
) -> None:
    """Test case: a non-security user cannot read automation state.

    Preparation:
    - The gateway is running and the general operator or restricted test user
      has a valid login without a security role.

    Action:
    - Request one Security Automation read endpoint as that user.

    Expected result:
    - The gateway denies access with HTTP 403.

    Requirement: security-automation/FR-009.
    """
    cookie = str(request.getfixturevalue(identity))
    response = api_client.get(path=path, cookies={ACCESS_COOKIE: cookie})
    expect(response.status).to_be(403)


@pytest.mark.parametrize(
    ("path", "payload"),
    [
        (
            f"{BASE}/rules/network_reconnaissance/mitigate",
            {"operands": {"requester_ip_address": "198.18.254.250"}},
        ),
        (f"{BASE}/rules/network_reconnaissance/enabled", {"enabled": True}),
        (f"{BASE}/active/12345678/rollback", {"reason": "authorization test"}),
    ],
)
@pytest.mark.parametrize("identity", ["operator_cookie", "testsudo_cookie"])
def test_operator_cannot_mutate_automation_api(
    api_client: RESTClient,
    request: pytest.FixtureRequest,
    identity: str,
    path: str,
    payload: dict[str, object],
) -> None:
    """Test case: a non-security user cannot invoke automation controls.

    Preparation:
    - The gateway has issued a valid general-operator or restricted-user session.

    Action:
    - Submit one well-formed automation mutation request.

    Expected result:
    - The gateway denies the request with HTTP 403 before sending the RPC.

    Requirement: security-automation/FR-009.
    """
    cookie = str(request.getfixturevalue(identity))
    if path.endswith("/enabled"):
        response = api_client.put(path=path, payload=payload, cookies={ACCESS_COOKIE: cookie})
    else:
        response = api_client.post(path=path, payload=payload, cookies={ACCESS_COOKIE: cookie})
    expect(response.status).to_be(403)


def test_security_operator_can_list_configured_rules(
    api_client: RESTClient, security_cookie: str
) -> None:
    """Test case: the security role can inspect packaged automation rules.

    Preparation:
    - The Security Automation service and gateway are running.

    Action:
    - Read the rule inventory as the security test account.

    Expected result:
    - The response contains the named network reconnaissance rule.

    Requirement: security-automation/FR-009, security-automation/FR-011.
    """
    response = api_client.get(path=f"{BASE}/rules", cookies={ACCESS_COOKIE: security_cookie})
    if response.status != HTTPStatus.OK:
        pytest.fail(f"automation rule list returned HTTP {response.status}")
    rules = response.body.get("data", {}).get("rules", [])
    expect(
        any(isinstance(rule, dict) and rule.get("id") == "network_reconnaissance" for rule in rules)
    ).to_be(True)


def test_force_mitigation_is_visible_in_security_alert_api(
    api_client: RESTClient,
    security_cookie: str,
    cli_client: CLIClient,
    source_ip_cleanup: str,
) -> None:
    """Test case: an applied automation effect emits a retrievable alert.

    Preparation:
    - The dedicated host has enrolled the network reconnaissance action.
    - The test source has no preexisting automation incident.

    Action:
    - Force mitigation through the authenticated CLI and read security events.

    Expected result:
    - A new security alert identifies this rule and source IP.

    Requirement: security-automation/FR-007, security-automation/FR-009.
    """
    path = "/api/alerts/security?page=1&size=500"
    before = api_client.get(path=path, cookies={ACCESS_COOKIE: security_cookie})
    if before.status != HTTPStatus.OK:
        pytest.fail(f"security alert list returned HTTP {before.status}")
    prior = before.body.get("data", {}).get("items", [])
    latest = max(
        (int(str(item.get("RecID", 0))) for item in prior if isinstance(item, dict)),
        default=0,
    )
    run_automation_cli(
        cli_client,
        SECURITY_LOGGING_MANAGER_LOGIN,
        SECURITY_LOGGING_MANAGER_PASSWORD,
        "force-mitigate",
        "network_reconnaissance",
        f"requester_ip_address={source_ip_cleanup}",
        "--yes",
    )
    found = False
    deadline = time.monotonic() + 30
    while time.monotonic() < deadline:
        response = api_client.get(path=path, cookies={ACCESS_COOKIE: security_cookie})
        if response.status != HTTPStatus.OK:
            pytest.fail(f"security alert list returned HTTP {response.status}")
        items = response.body.get("data", {}).get("items", [])
        found = any(
            isinstance(item, dict)
            and int(str(item.get("RecID", 0))) > latest
            and item.get("Source") == "security-automation"
            and source_ip_cleanup in str(item.get("Message", ""))
            for item in items
        )
        if found:
            break
        time.sleep(0.5)
    expect(found).to_be(True)
