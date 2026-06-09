"""Shared UI system-test fixtures for the LiteNAS admin panel."""

from collections.abc import Generator

from constants import UI_BASE_URL, UI_BROWSER_CAPS
from hyperiontf.executors.pytest import fixture
from ui.page_objects.login_page import LoginPage


@fixture(scope="function", log=False)  # type: ignore[untyped-decorator]
def login_page() -> Generator[LoginPage, None, None]:
    """Open the admin-panel root route in a fresh anonymous browser session."""
    page = LoginPage.start_browser(UI_BROWSER_CAPS)
    page.open(UI_BASE_URL)
    yield page
    page.quit()
