"""Installed-system Security Journal browser workflow."""

import pytest
from constants import SECURITY_LOGGING_MANAGER_LOGIN, SECURITY_LOGGING_MANAGER_PASSWORD, UI_BASE_URL
from hyperiontf.executors.pytest import hyperion_test_case_setup  # noqa: F401
from ui.page_objects.journal_page import JournalPage
from ui.page_objects.login_page import LoginPage


@pytest.mark.JournalManager
@pytest.mark.ui
def test_security_user_can_open_journal_browser(login_page: LoginPage) -> None:
    """Test case: a security user can open the journal modal.

    Preparation:
    - The installed journal manager, gateway, and admin panel are running.
    - The security test user can sign in.

    Action:
    - Sign in, navigate to Security Journal, and open its modal.

    Expected result:
    - The journal modal exposes its search control.
    """
    login_page.sign_in(SECURITY_LOGGING_MANAGER_LOGIN, SECURITY_LOGGING_MANAGER_PASSWORD)
    page = JournalPage(login_page.automation_adapter)
    page.open(f"{UI_BASE_URL}/security/journal")
    dialog = page.browse()
    dialog.search_field.assert_visible()
