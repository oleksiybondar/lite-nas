"""Installed gateway JSON body validation for security actions."""

from http import HTTPStatus

import pytest
from hyperiontf import RESTClient, expect

pytestmark = [pytest.mark.api, pytest.mark.WebGateway]


# Requirements: security-monitoring/FR-008.
@pytest.mark.Antivirus
def test_antivirus_scan_rejects_empty_folder_body(
    api_client: RESTClient, security_cookie: str
) -> None:
    """Test case: a scan request needs a valid folder in its JSON body.

    Preparation:
    - The security test user is authenticated and Antivirus is running.

    Action:
    - Submit an empty folder to the scan start endpoint.

    Expected result:
    - Gateway validation rejects the request before a scan is queued.
    """
    response = api_client.post(
        path="/api/security/tasks/antivirus/scans",
        payload={"folder": ""},
        cookies={"lite-nas-at": security_cookie},
    )
    expect(response.status).to_be(HTTPStatus.UNPROCESSABLE_ENTITY)


# Requirements: usb-manager/FR-002.
@pytest.mark.USBManager
def test_usbguard_action_rejects_nonboolean_permanent_body(
    api_client: RESTClient, security_cookie: str
) -> None:
    """Test case: USBGuard policy actions require a Boolean JSON decision.

    Preparation:
    - The security test user is authenticated; no test USB device is selected.

    Action:
    - Submit a string instead of a Boolean for an absent device ID.

    Expected result:
    - Gateway validation rejects the action before USBGuard is called.
    """
    response = api_client.post(
        path="/api/usbguard/devices/9999999999/allow",
        payload={"permanent": "yes"},
        cookies={"lite-nas-at": security_cookie},
    )
    expect(response.status).to_be(HTTPStatus.UNPROCESSABLE_ENTITY)
