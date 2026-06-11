"""Expanded sidebar shell widget for authenticated admin-panel pages."""

from __future__ import annotations

from hyperiontf import By, Widget, widgets
from hyperiontf.ui.elements import Elements
from ui.page_objects.widgets.sidebar.item import SidebarNavigationItemWidget


class AppSidebarWidget(Widget):  # type: ignore[misc]
    """Expanded desktop sidebar region used for authenticated dashboard navigation."""

    @widgets(klass=SidebarNavigationItemWidget)  # type: ignore[untyped-decorator]
    def items(self) -> Elements:
        """Top-level sidebar rows rendered inside the dashboard navigation tree."""
        return By.css("[data-testid='app-sidebar-list'] > [data-test-class='sidebar-tree-item']")
