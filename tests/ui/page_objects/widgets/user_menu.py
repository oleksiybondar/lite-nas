"""Authenticated user action menu widgets."""

from hyperiontf import By, Widget, element
from hyperiontf.ui.element import Element


class UserMenuWidget(Widget):  # type: ignore[misc]
    """Detached menu that exposes authenticated user account actions."""

    @element  # type: ignore[untyped-decorator]
    def profile_link(self) -> Element:
        """Menu item that navigates to the current user's profile page."""
        return By.css("[data-testid='user-menu-profile-link']").from_document()

    @element  # type: ignore[untyped-decorator]
    def application_settings_link(self) -> Element:
        """Menu item that navigates to application settings."""
        return By.css("[data-testid='user-menu-application-settings-link']").from_document()

    @element  # type: ignore[untyped-decorator]
    def logout_button(self) -> Element:
        """Menu item that logs out the current authenticated session."""
        return By.css("[data-testid='user-menu-logout-button']").from_document()

    def open_profile(self) -> None:
        """Navigate to the authenticated user's profile page from the menu."""
        self.profile_link.click()

    def open_application_settings(self) -> None:
        """Navigate to application settings from the menu."""
        self.application_settings_link.click()

    def logout(self) -> None:
        """Log out the authenticated session from the menu."""
        self.logout_button.click()
