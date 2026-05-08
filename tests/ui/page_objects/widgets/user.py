"""Authenticated user widgets for the admin-panel top bar."""

from hyperiontf import By, Widget, element
from hyperiontf.ui.element import Element


class CurrentUserWidget(Widget):  # type: ignore[misc]
    """Authenticated user summary control shown in the application top bar."""

    @element  # type: ignore[untyped-decorator]
    def avatar(self) -> Element:
        """Avatar image or initials for the authenticated user."""
        return By.css("[data-testid='user-avatar']")

    @element  # type: ignore[untyped-decorator]
    def summary(self) -> Element:
        """Visible user identity summary grouped with the avatar."""
        return By.css("[data-testid='user-menu-summary']")

    @element  # type: ignore[untyped-decorator]
    def login(self) -> Element:
        """Login name displayed for the authenticated user."""
        return By.css("[data-testid='user-menu-login']")

    @element  # type: ignore[untyped-decorator]
    def full_name(self) -> Element:
        """Optional full name displayed below the authenticated user's login."""
        return By.css("[data-testid='user-menu-full-name']")

    def open_menu(self) -> None:
        """Open the authenticated user action menu from the user summary button."""
        self.click()

    def assert_authenticated_login(self, expected_login: str) -> None:
        """Assert that the top bar shows the authenticated user's login name."""
        self.assert_visible()
        self.login.assert_text(expected_login)
