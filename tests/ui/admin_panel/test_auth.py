"""System UI test suite for browser-facing authentication flows."""

import pytest
from constants import CREDENTIALS
from hyperiontf.executors.pytest import hyperion_test_case_setup  # noqa: F401
from ui.page_objects.dashboard_page import DashboardPage
from ui.page_objects.login_page import LoginPage


@pytest.mark.Auth
@pytest.mark.ui
def test_anonymous_browser_renders_login_page(login_page: LoginPage) -> None:
    """Test case: anonymous browser renders the login page.

    Preparation:
    - A fresh browser session has no authenticated LiteNAS admin-panel cookies.

    Action:
    - Open the admin-panel root route.

    Expected result:
    - The browser renders the login page.
    """
    login_page.login_form.wait_until_found()
    login_page.login_form.assert_visible()


@pytest.mark.Auth
@pytest.mark.ui
def test_successful_login_renders_authenticated_user(login_page: LoginPage) -> None:
    """Test case: successful login renders the authenticated user in the top bar.

    Preparation:
    - A fresh browser session has loaded the login page.
    - The configured test user exists with the configured password.

    Action:
    - Submit the configured credentials through the login form.

    Expected result:
    - The top bar shows the configured authenticated login name.
    """
    login_page.sign_in(CREDENTIALS["login"], CREDENTIALS["password"])
    dashboard_page = DashboardPage(login_page.automation_adapter)
    dashboard_page.top_bar.actions.user.assert_authenticated_login(CREDENTIALS["login"])
