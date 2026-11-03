"""HyperionTF model of the Moments sign-in form."""

from hyperiontf import By, Widget, element
from hyperiontf.ui.element import Element


class MomentsLoginFormWidget(Widget):  # type: ignore[misc]
    """Moments credential form that submits a browser session for a Unix user."""

    @element  # type: ignore[untyped-decorator]
    def login_field(self) -> Element:
        """Login input inside the Moments credential form."""
        return By.css("[data-testid='moments-login-name'] input")

    @element  # type: ignore[untyped-decorator]
    def password_field(self) -> Element:
        """Password input inside the Moments credential form."""
        return By.css("[data-testid='moments-login-password'] input")

    @element  # type: ignore[untyped-decorator]
    def submit_button(self) -> Element:
        """Button that sends the entered credentials to Moments Auth."""
        return By.css("[data-testid='moments-login-submit']")

    def sign_in(self, login: str, password: str) -> None:
        """Enter credentials and submit the form to establish a browser session."""
        self.login_field.clear_and_fill(login)
        self.password_field.clear_and_fill(password)
        self.submit_button.click()
