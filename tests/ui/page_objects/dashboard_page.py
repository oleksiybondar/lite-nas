"""Authenticated admin dashboard shell page objects."""

from hyperiontf import By, WebPage, widget
from ui.page_objects.widgets.sidebar import AppSidebarWidget
from ui.page_objects.widgets.top_bar import AppTopBarWidget


class DashboardPage(WebPage):  # type: ignore[misc]
    """Authenticated admin-panel dashboard shell with top navigation and sidebar."""

    @widget(klass=AppTopBarWidget)  # type: ignore[untyped-decorator]
    def top_bar(self) -> AppTopBarWidget:
        """Top navigation bar containing app branding and authenticated user actions."""
        return By.css("[data-testid='app-top-bar']")  # type: ignore[no-any-return]

    @widget(klass=AppSidebarWidget)  # type: ignore[untyped-decorator]
    def sidebar(self) -> AppSidebarWidget:
        """Expanded desktop sidebar navigation shown in non-collapsed dashboard mode."""
        return By.css("[data-testid='app-sidebar']")  # type: ignore[no-any-return]
