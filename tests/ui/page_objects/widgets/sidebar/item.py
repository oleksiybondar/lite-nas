"""Sidebar navigation row widgets for authenticated admin-panel pages."""

from __future__ import annotations

from hyperiontf import By, Widget, element, widgets
from hyperiontf.ui.element import Element
from hyperiontf.ui.elements import Elements


class SidebarNavigationItemWidget(Widget):  # type: ignore[misc]
    """One clickable sidebar row that may own nested child navigation rows."""

    @element  # type: ignore[untyped-decorator]
    def label(self) -> Element:
        """Visible label text for this sidebar navigation row."""
        return By.css("[data-test-class='sidebar-tree-item-label']")

    @element  # type: ignore[untyped-decorator]
    def icon(self) -> Element:
        """Optional icon cell rendered before the sidebar row label."""
        return By.css("[data-test-class='sidebar-tree-item-icon']")

    @element  # type: ignore[untyped-decorator]
    def expand_control(self) -> Element:
        """Expand or collapse control for sidebar rows with child routes."""
        return By.css("[data-test-class='sidebar-tree-expand-control']")

    @widgets(klass=lambda: SidebarNavigationItemWidget)  # type: ignore[untyped-decorator]
    def children(self) -> Elements:
        """Nested sidebar rows rendered under this branch when it is expanded."""
        return By.xpath(
            "./following-sibling::*[@data-test-class='sidebar-tree-children']//*[@data-test-class='sidebar-tree-item']"
        )

    def navigate(self) -> None:
        """Navigate to the route represented by this sidebar row."""
        self.click()

    def expand(self) -> None:
        """Expand this sidebar row when it owns nested child routes."""
        self.expand_control.click()

    def collapse(self) -> None:
        """Collapse this sidebar row when its child route list is expanded."""
        self.expand_control.click()
