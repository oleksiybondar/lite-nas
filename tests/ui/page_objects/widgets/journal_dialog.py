"""Journal modal widget for installed admin-panel browser tests."""

from hyperiontf import By, Widget, element
from hyperiontf.ui.element import Element


class JournalDialogWidget(Widget):  # type: ignore[misc]
    """Modal containing journal filters, entries, and chunk navigation."""

    @element  # type: ignore[untyped-decorator]
    def search_field(self) -> Element:
        """Message search input owned by the journal modal."""
        return By.css("[data-testid='journal-search'] input")

    @element  # type: ignore[untyped-decorator]
    def apply_filters_button(self) -> Element:
        """Apply current modal filters and replace the displayed chunk."""
        return By.css("[data-testid='journal-apply-filters']")

    @element  # type: ignore[untyped-decorator]
    def load_more_button(self) -> Element:
        """Append the next journal chunk when one is available."""
        return By.css("[data-testid='journal-load-more']")

    def search_message(self, text: str) -> None:
        """Set literal message search and request its first result chunk."""
        self.search_field.clear_and_fill(text)
        self.apply_filters_button.click()
