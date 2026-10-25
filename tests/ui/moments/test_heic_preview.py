"""Installed-system HEIC browser rendering through HyperionTF page objects."""

import pytest
from constants import (
    SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
    SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
    UI_BASE_URL,
    UI_BROWSER_CAPS,
)
from hyperiontf.executors.pytest import hyperion_test_case_setup  # noqa: F401
from ui.page_objects.moments_home_page import MomentsHomePage
from ui.page_objects.moments_login_page import MomentsLoginPage
from ui.page_objects.stream_preview_page import StreamPreviewPage

pytestmark = [pytest.mark.ui, pytest.mark.Moments, pytest.mark.StreamService]
SYNTHETIC_HEIC_SIZE = 64


def test_browser_decodes_authorized_heic_preview(operator_heic: str) -> None:
    """Test case: a logged-in browser displays a private HEIC as a JPEG preview.

    Preparation:
    - A synthetic HEIC belongs to the operator and Moments is installed.

    Action:
    - Sign in through Moments and open the HEIC preview URL in the same browser.

    Expected result:
    - The browser decodes the authorized preview image.
    """
    login_page = MomentsLoginPage.start_browser(UI_BROWSER_CAPS)
    try:
        login_page.open(f"{UI_BASE_URL.rstrip('/')}/apps/moments/")
        login_page.sign_in(
            SYSTEM_LOGGING_MANAGER_OPERATOR_LOGIN,
            SYSTEM_LOGGING_MANAGER_OPERATOR_PASSWORD,
        )
        home_page = MomentsHomePage(login_page.automation_adapter)
        home_page.gallery.wait_until_found()
        preview_page = StreamPreviewPage(login_page.automation_adapter)
        preview_page.open(f"{UI_BASE_URL.rstrip('/')}/s/preview{operator_heic}")
        assert preview_page.image_rendered(SYNTHETIC_HEIC_SIZE)
    finally:
        login_page.quit()
