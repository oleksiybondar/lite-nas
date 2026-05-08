"""Top navigation bar widgets for LiteNAS admin-panel pages."""

from hyperiontf import By, Widget, element, widget
from hyperiontf.ui.element import Element
from ui.page_objects.widgets.user import CurrentUserWidget
from ui.page_objects.widgets.user_menu import UserMenuWidget


class AppTopBarActionsWidget(Widget):  # type: ignore[misc]
    """Trailing top-bar action region that owns authenticated user controls."""

    @widget(klass=CurrentUserWidget)  # type: ignore[untyped-decorator]
    def user(self) -> CurrentUserWidget:
        """Authenticated user summary button rendered in the actions region."""
        return By.css("[data-testid='user-menu-button']")  # type: ignore[no-any-return]

    def open_user_menu(self) -> None:
        """Open the authenticated user action menu from the actions region."""
        self.user.open_menu()


class AppTopBarWidget(Widget):  # type: ignore[misc]
    """Application top bar shared by public and authenticated layouts."""

    @element  # type: ignore[untyped-decorator]
    def toolbar(self) -> Element:
        """Toolbar area containing branding, navigation controls, and actions."""
        return By.css("[data-testid='app-top-bar-toolbar']")

    @widget(klass=AppTopBarActionsWidget)  # type: ignore[untyped-decorator]
    def actions(self) -> AppTopBarActionsWidget:
        """Trailing action region that hosts authenticated user controls."""
        return By.css("[data-testid='app-top-bar-actions']")  # type: ignore[no-any-return]

    @widget(klass=UserMenuWidget)  # type: ignore[untyped-decorator]
    def user_menu(self) -> UserMenuWidget:
        """Detached authenticated user action menu opened from the top bar."""
        return By.css("[data-testid='user-menu']").from_document()  # type: ignore[no-any-return]

    def open_user_menu(self) -> UserMenuWidget:
        """Open and return the authenticated user action menu."""
        self.actions.open_user_menu()
        return self.user_menu  # type: ignore[no-any-return]
