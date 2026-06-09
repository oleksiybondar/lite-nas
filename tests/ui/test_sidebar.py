"""System UI test suite for sidebar RBAC visibility on the admin panel."""

import pytest
from constants import (
    CREDENTIALS,
    SECURITY_LOGGING_MANAGER_LOGIN,
    SECURITY_LOGGING_MANAGER_PASSWORD,
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
    SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
)
from hyperiontf.executors.pytest import hyperion_test_case_setup  # noqa: F401
from ui.page_objects.dashboard_page import DashboardPage
from ui.page_objects.login_page import LoginPage
from ui.page_objects.widgets.sidebar.item import SidebarNavigationItemWidget


@pytest.mark.Auth
@pytest.mark.ui
@pytest.mark.parametrize(
    "case",
    [
        pytest.param(
            (
                CREDENTIALS["login"],
                CREDENTIALS["password"],
                (False, False, False),
            ),
            id="testuser",
        ),
        pytest.param(
            (
                SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
                SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
                (True, True, False),
            ),
            id="testoperator",
        ),
        pytest.param(
            (
                SECURITY_LOGGING_MANAGER_LOGIN,
                SECURITY_LOGGING_MANAGER_PASSWORD,
                (True, False, True),
            ),
            id="testsecurity",
        ),
    ],
)
def test_sidebar_alerts_rbac_visibility(
    login_page: LoginPage,
    case: tuple[str, str, tuple[bool, bool, bool]],
) -> None:
    """Test case: sidebar alert navigation matches the authenticated user's RBAC scope.

    Preparation:
    - A fresh browser session has loaded the admin-panel login page.
    - The parametrized system-test user exists with the configured password.

    Action:
    - Sign in as the parametrized user and inspect the Alerts branch in the sidebar.

    Expected result:
    - The sidebar hides the Alerts branch for the restricted user.
    - The sidebar shows Alerts and the expected role-specific child items for
      operator and security users.
    """
    login, password, (has_alerts, has_system, has_security) = case
    login_page.sign_in(login, password)
    dashboard_page = DashboardPage(login_page.automation_adapter)
    sidebar = dashboard_page.sidebar
    sidebar.wait_until_found()

    alerts_item = sidebar.items['text == "Alerts"']
    _assert_item_visibility(alerts_item, has_alerts)
    if not has_alerts:
        return

    children = alerts_item.expand()
    _assert_item_visibility(children['text == "System"'], has_system)
    _assert_item_visibility(children['text == "Security"'], has_security)


def _assert_item_visibility(
    item: SidebarNavigationItemWidget | None,
    is_present: bool,
) -> None:
    """Assert that one sidebar row is either rendered and visible or absent."""
    if is_present:
        assert item is not None
        item.assert_visible()
        return

    assert item is None
