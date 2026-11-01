"""Security Journal page object for installed admin-panel browser tests."""

from typing import cast

from hyperiontf import By, WebPage, element, widget
from hyperiontf.ui.element import Element
from ui.page_objects.widgets.journal_dialog import JournalDialogWidget


class JournalPage(WebPage):  # type: ignore[misc]
    """Security page that owns journal browsing and vacuum controls."""

    @element  # type: ignore[untyped-decorator]
    def browse_button(self) -> Element:
        """Open the journal browsing modal from the Security Journal page."""
        return By.css("[data-testid='journal-browse']")

    @widget(klass=JournalDialogWidget)  # type: ignore[untyped-decorator]
    def dialog(self) -> JournalDialogWidget:
        """Journal modal composed by the Security Journal page."""
        return By.css("[data-testid='journal-dialog']")  # type: ignore[no-any-return]

    def browse(self) -> JournalDialogWidget:
        """Open the journal modal and wait for its controls to appear."""
        self.browse_button.click()
        self.dialog.wait_until_found()
        return cast(JournalDialogWidget, self.dialog)
