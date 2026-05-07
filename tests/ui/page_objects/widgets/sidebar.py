"""Sidebar navigation widgets for authenticated admin-panel pages."""

from typing import cast

from hyperiontf import By, Widget, element, elements, widget, widgets
from hyperiontf.ui.element import Element
from hyperiontf.ui.elements import Elements


class SidebarNavigationItemWidget(Widget):  # type: ignore[misc]
    """One clickable row in the expanded dashboard sidebar tree."""

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

    def navigate(self) -> None:
        """Navigate to the route represented by this sidebar row."""
        self.click()

    def expand(self) -> None:
        """Expand this sidebar row when it owns nested child routes."""
        self.expand_control.click()

    def collapse(self) -> None:
        """Collapse this sidebar row when its child route list is expanded."""
        self.expand_control.click()


class SidebarNavigationTreeWidget(Widget):  # type: ignore[misc]
    """Expanded sidebar navigation tree that owns route rows and branches."""

    @widgets(klass=SidebarNavigationItemWidget)  # type: ignore[untyped-decorator]
    def items(self) -> Elements:
        """Currently rendered sidebar navigation rows."""
        return By.css("[data-test-class='sidebar-tree-item']")

    @elements  # type: ignore[untyped-decorator]
    def expanded_child_lists(self) -> Elements:
        """Rendered child-route lists for expanded sidebar branches."""
        return By.css("[data-test-class='sidebar-tree-children']")

    def item_by_name(self, name: str) -> SidebarNavigationItemWidget:
        """Return the currently rendered sidebar row with the given visible name."""
        return self._item_by_attribute("data-test-name", name)

    def item_by_path(self, path: str) -> SidebarNavigationItemWidget:
        """Return the currently rendered sidebar row for the given route path."""
        return self._item_by_attribute("data-test-path", path)

    def navigate_to_name(self, name: str) -> None:
        """Navigate through the currently rendered sidebar row with the given name."""
        self.item_by_name(name).navigate()

    def navigate_to_path(self, path: str) -> None:
        """Navigate through the currently rendered sidebar row with the given path."""
        self.item_by_path(path).navigate()

    def expand_item(self, name: str) -> None:
        """Expand the currently rendered sidebar row with the given visible name."""
        self.item_by_name(name).expand()

    def collapse_item(self, name: str) -> None:
        """Collapse the currently rendered sidebar row with the given visible name."""
        self.item_by_name(name).collapse()

    def _item_by_attribute(self, attribute_name: str, value: str) -> SidebarNavigationItemWidget:
        """Resolve one rendered sidebar item by a semantic test attribute."""
        query = f'attribute:{attribute_name} == "{self._escape_eql_value(value)}"'
        item = self.items[query]
        if item is None:
            msg = f"Sidebar item with {attribute_name}={value!r} is not currently rendered."
            raise AssertionError(msg)

        return cast(SidebarNavigationItemWidget, item)

    def _escape_eql_value(self, value: str) -> str:
        """Escape a string value for use inside a double-quoted EQL literal."""
        return value.replace("\\", "\\\\").replace('"', '\\"')


class AppSidebarWidget(Widget):  # type: ignore[misc]
    """Expanded desktop sidebar region used for authenticated dashboard navigation."""

    @widget(klass=SidebarNavigationTreeWidget)  # type: ignore[untyped-decorator]
    def tree(self) -> SidebarNavigationTreeWidget:
        """Navigation tree containing expanded sidebar route rows."""
        return By.css("[data-testid='app-sidebar-list']")  # type: ignore[no-any-return]

    def item_by_name(self, name: str) -> SidebarNavigationItemWidget:
        """Return the currently rendered sidebar row with the given visible name."""
        return cast(SidebarNavigationItemWidget, self.tree.item_by_name(name))

    def item_by_path(self, path: str) -> SidebarNavigationItemWidget:
        """Return the currently rendered sidebar row for the given route path."""
        return cast(SidebarNavigationItemWidget, self.tree.item_by_path(path))

    def navigate_to_name(self, name: str) -> None:
        """Navigate through the currently rendered sidebar row with the given name."""
        self.tree.navigate_to_name(name)

    def navigate_to_path(self, path: str) -> None:
        """Navigate through the currently rendered sidebar row with the given path."""
        self.tree.navigate_to_path(path)

    def expand_item(self, name: str) -> None:
        """Expand the currently rendered sidebar row with the given visible name."""
        self.tree.expand_item(name)

    def collapse_item(self, name: str) -> None:
        """Collapse the currently rendered sidebar row with the given visible name."""
        self.tree.collapse_item(name)
