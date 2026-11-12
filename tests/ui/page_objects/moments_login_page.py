"""HyperionTF page model for anonymous Moments sessions."""

from hyperiontf import By, WebPage, widget
from ui.page_objects.widgets.moments_login_form import MomentsLoginFormWidget


class MomentsLoginPage(WebPage):  # type: ignore[misc]
    """Anonymous Moments page whose sign-in form creates the media cookie."""

    @widget(klass=MomentsLoginFormWidget)  # type: ignore[untyped-decorator]
    def login_form(self) -> MomentsLoginFormWidget:
        """Credential form shown before the authenticated Moments gallery."""
        return By.css("[data-testid='moments-login']")  # type: ignore[no-any-return]

    def sign_in(self, login: str, password: str) -> None:
        """Wait for the form, then establish an authenticated Moments session."""
        self.login_form.wait_until_found()
        self.login_form.sign_in(login, password)
