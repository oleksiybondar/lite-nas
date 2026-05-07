"""Login form widget for anonymous admin-panel sessions."""

from hyperiontf import By, Widget, element
from hyperiontf.ui.element import Element


class LoginFormWidget(Widget):  # type: ignore[misc]
    """Credential form that owns the anonymous sign-in interaction."""

    @element  # type: ignore[untyped-decorator]
    def title(self) -> Element:
        """Visible login form heading shown above credential fields."""
        return By.css("[data-testid='login-title']")

    @element  # type: ignore[untyped-decorator]
    def username_field(self) -> Element:
        """Login-name input field for local usernames or email addresses."""
        return By.css("[data-testid='login-username-field'] input")

    @element  # type: ignore[untyped-decorator]
    def password_field(self) -> Element:
        """Password input field for the submitted login credentials."""
        return By.css("[data-testid='login-password-field'] input")

    @element  # type: ignore[untyped-decorator]
    def submit_button(self) -> Element:
        """Primary button that submits the login attempt."""
        return By.css("[data-testid='login-submit-button']")

    @element  # type: ignore[untyped-decorator]
    def error_message(self) -> Element:
        """Inline error message displayed after an unsuccessful login attempt."""
        return By.css("[data-testid='login-error-message']")

    def sign_in(self, login: str, password: str) -> None:
        """Submit credentials through this login form."""
        self.enter_credentials(login, password)
        self.submit()

    def enter_credentials(self, login: str, password: str) -> None:
        """Replace the current credential fields with the provided values."""
        self.username_field.clear_and_fill(login)
        self.password_field.clear_and_fill(password)

    def submit(self) -> None:
        """Submit the login form with the primary sign-in button."""
        self.submit_button.click()
