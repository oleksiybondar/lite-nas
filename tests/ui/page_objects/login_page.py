"""Login page object for anonymous admin-panel sessions."""

from hyperiontf import By, WebPage, widget
from ui.page_objects.widgets.login_form import LoginFormWidget
from ui.page_objects.widgets.top_bar import AppTopBarWidget


class LoginPage(WebPage):  # type: ignore[misc]
    """Anonymous sign-in page with public chrome and a credential form."""

    @widget(klass=AppTopBarWidget)  # type: ignore[untyped-decorator]
    def top_bar(self) -> AppTopBarWidget:
        """Public top navigation bar rendered above the login form."""
        return By.css("[data-testid='app-top-bar']")  # type: ignore[no-any-return]

    @widget(klass=LoginFormWidget)  # type: ignore[untyped-decorator]
    def login_form(self) -> LoginFormWidget:
        """Credential form region that owns login fields and submission."""
        return By.css("[data-testid='login-form']")  # type: ignore[no-any-return]

    def sign_in(self, login: str, password: str) -> None:
        """Submit credentials through the page's login form."""
        self.login_form.wait_until_found()
        self.login_form.sign_in(login, password)
